package bit

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"axlab.dev/bit/files"
	"axlab.dev/bit/logs"
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
	logs.Out("\n○○○ %s: compiling from at `%s` to `%s`...\n", header, input.Name(), comp.buildDir.Name())

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
		logs.Out("\n>>> Queued `%s`...\n", inputPath)
		program.compiler.pending.Add(1)
		go func() {
			defer close(wait)
			defer program.compiler.pending.Done()
			program.buildMutex.Lock()
			defer program.compiling.Store(false)
			defer program.buildMutex.Unlock()

			startTime := time.Now()
			defer func() {
				duration := time.Since(startTime)
				logs.Out("<<< Finished `%s` in %s\n", inputPath, duration.String())
			}()

			compiler := program.compiler
			if source, err := compiler.LoadSource(inputPath); err == nil {
				logs.Out("... Compiling `%s`...\n", inputPath)
				program.Compile(source)
			} else {
				logs.Err("!!! Failed to load `%s`: %v", inputPath, err)
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

		logs.Out("\n>>> Removing `%s`\n", program.config.InputPath)
		program.compiler.buildDir.RemoveAll(program.config.BuildPath)
	}()
}
