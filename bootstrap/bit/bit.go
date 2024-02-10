package bit

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"axlab.dev/bit/common"
	"axlab.dev/bit/files"
	"axlab.dev/bit/proc"
)

const MaxErrorOutput = 16

const (
	PrecFirst Precedence = iota
	PrecBrackets
	PrecIndent
	PrecLines
	PrecPrint
	PrecReplace
	PrecOutput
	PrecLast
)

type Precedence int

type Compiler struct {
	ctx context.Context

	inputDir files.Dir
	buildDir files.Dir

	programsMutex sync.Mutex
	programs      map[string]map[string]*Program

	pending sync.WaitGroup

	sourceFileMutex sync.Mutex
	sourceFileMap   map[string]*struct {
		src *Source
		err error
	}
}

func NewCompiler(ctx context.Context, inputPath, buildPath string) *Compiler {
	return &Compiler{
		ctx:      ctx,
		inputDir: files.OpenDir(inputPath).MustExist("compiler input"),
		buildDir: files.OpenDir(buildPath).Create("compiler build"),
	}
}

func (comp *Compiler) InputDir() files.Dir {
	return comp.inputDir
}

func (comp *Compiler) BuildDir() files.Dir {
	return comp.buildDir
}

func (comp *Compiler) Watch(once bool) {
	input := comp.inputDir
	inputPath := input.FullPath()

	watcher := files.Watch(inputPath, files.ListOptions{Filter: func(entry *files.Entry) bool {
		if entry.IsDir {
			return true
		}
		return strings.HasSuffix(entry.Name, ".bit")
	}})

	inputEvents := watcher.Start(100 * time.Millisecond)

	header := "Watcher"
	if once {
		header = "Build"
	}
	common.Out("\n○○○ %s: compiling from at `%s` to `%s`...\n", header, input.Name(), comp.buildDir.Name())

	var programs []*Program
	for _, it := range watcher.List() {
		if it.IsDir {
			continue
		}

		buildPath := it.Path
		program := comp.GetProgram(it.Path, buildPath)
		programs = append(programs, program)

		force := once
		program.QueueCompile(force)
	}

mainLoop:
	for !once {
		select {
		case events := <-inputEvents:
			comp.FlushSources()
			for _, ev := range events {
				if ev.Entry.IsDir {
					continue
				}

				buildPath := ev.Entry.Path
				program := comp.GetProgram(ev.Entry.Path, buildPath)
				if ev.Type != files.EventRemove {
					program.QueueCompile(false)
				} else {
					program.ClearBuild()
				}
			}
		case <-comp.ctx.Done():
			break mainLoop
		}
	}

	comp.pending.Wait()
	if once {
		for _, it := range programs {
			if len(it.errors) > 0 {
				os.Stderr.WriteString(fmt.Sprintf("\n>>> Program %s <<<\n\n", it.source.Name()))
			}
			it.ShowErrors()
		}
	}
}

func (comp *Compiler) GetProgram(rootFile, outputDir string) *Program {
	inputPath, inputName := comp.inputDir.ResolvePath(rootFile)
	buildPath, buildName := comp.buildDir.ResolvePath(outputDir)

	comp.programsMutex.Lock()
	defer comp.programsMutex.Unlock()
	if comp.programs == nil {
		comp.programs = make(map[string]map[string]*Program)
	}

	outputMap := comp.programs[inputPath]
	if outputMap == nil {
		outputMap = make(map[string]*Program)
		comp.programs[inputPath] = outputMap
	}

	program := outputMap[buildPath]
	if program == nil {
		program = NewProgram(comp, ProgramConfig{
			InputPath: inputName,
			BuildPath: buildName,
		})
		program.InitCore()
		outputMap[buildPath] = program
	}

	return program
}

func (program *Program) QueueCompile(force bool) (wait chan struct{}) {
	wait = make(chan struct{})
	recompile := force || program.NeedRecompile()
	if recompile && program.compiling.CompareAndSwap(false, true) {
		inputPath := program.config.InputPath
		common.Out("\n>>> Queued `%s`...\n", inputPath)
		program.compiler.pending.Add(1)
		go func() {
			defer close(wait)
			defer program.compiler.pending.Done()
			program.buildMutex.Lock()
			defer program.compiling.Store(false)
			defer program.buildMutex.Unlock()

			startTime := time.Now()
			outputDuration := func(header string) {
				duration := time.Since(startTime)
				common.Out("%s%s\n", header, duration.String())
			}
			defer outputDuration(fmt.Sprintf("<<< Finished `%s` in ", inputPath))

			compiler := program.compiler
			if source, err := compiler.LoadSource(inputPath); err == nil {
				common.Out("... Compiling `%s`...\n", inputPath)
				program.Compile(source)
				outputDuration("... Compilation took ")

				if program.Valid() {
					common.Out("... Generating C output...\n")
					if ok, main := program.OutputCpp(); ok {
						outputCpp := program.outputPath("main.exe")
						outputCppFull := compiler.buildDir.GetFullPath(outputCpp)
						mainPath := compiler.buildDir.GetFullPath(main)
						common.Out("... Compiling C output to `%s`...\n", outputCpp)

						if !proc.Run("CC", "gcc", mainPath, "-o", outputCppFull) {
							common.Out("\nCompilation failed\n")
						}
					}
				}

				const resultFile = "result.txt"
				if program.Valid() {
					common.Out("... Running program...\n")
					rt := NewRuntime(program.mainNode)

					out := strings.Builder{}
					err := strings.Builder{}
					rt.StdOut = &out
					rt.StdErr = &err
					result := rt.Eval(*program.outputCode)

					res := strings.Builder{}
					res.WriteString("Result = ")
					res.WriteString(ResultRepr(result))
					res.WriteString("\n")

					if out.Len() > 0 {
						txt := out.String()
						res.WriteString("\n----- STDOUT -----\n\n")
						res.WriteString(txt)
						if c := txt[len(txt)-1]; c != '\n' && c != '\r' {
							res.WriteString("↵\n")
						}
					}

					if err.Len() > 0 {
						txt := err.String()
						res.WriteString("\n----- STDERR -----\n\n")
						res.WriteString(txt)
						if c := txt[len(txt)-1]; c != '\n' && c != '\r' {
							res.WriteString("↵\n")
						}
					}

					program.writeOutput(resultFile, res.String(), true)
				} else {
					program.removeOutput(resultFile)
				}
			} else {
				common.Err("!!! Failed to load `%s`: %v", inputPath, err)
			}
		}()
	} else {
		close(wait)
	}
	return wait
}

func (program *Program) ClearBuild() {
	go func() {
		program.buildMutex.Lock()
		defer program.compiling.Store(false)
		defer program.buildMutex.Unlock()

		common.Out("\n>>> Removing `%s`\n", program.config.InputPath)
		program.compiler.buildDir.RemoveAll(program.config.BuildPath)
	}()
}
