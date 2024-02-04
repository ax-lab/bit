package bit

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"axlab.dev/bit/files"
	"axlab.dev/bit/logs"
)

type CompilerError struct {
	Location Location
	Span     Span
	Message  string
	Args     []any
}

func (err CompilerError) String() string {
	msg := err.Message
	if len(err.Args) > 0 {
		msg = fmt.Sprintf(msg, err.Args)
	}
	loc := fmt.Sprintf("%s:%s", err.Span.Source().Name(), err.Location.String())
	txt := err.Span.DisplayText(0)
	if len(txt) > 0 {
		txt = fmt.Sprintf("\n\n    | %s", txt)
	}
	return fmt.Sprintf("at %s: %s%s", loc, msg, txt)
}

func (err CompilerError) Error() string {
	return err.String()
}

type Compiler struct {
	ctx context.Context

	inputDir files.Dir
	buildDir files.Dir

	programsMutex sync.Mutex
	programs      map[string]map[string]*Program

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

func (comp *Compiler) Watch() {
	input := comp.inputDir
	inputPath := input.FullPath()

	watcher := files.Watch(inputPath, files.ListOptions{Filter: func(entry *files.Entry) bool {
		if entry.IsDir {
			return true
		}
		return strings.HasSuffix(entry.Name, ".bit")
	}})

	inputEvents := watcher.Start(100 * time.Millisecond)

	logs.Out("\n○○○ Watcher: compiling from at `%s` to `%s`...\n", input.Name(), comp.buildDir.Name())

	for _, it := range watcher.List() {
		if it.IsDir {
			continue
		}

		buildPath := it.Path
		program := comp.GetProgram(it.Path, buildPath)
		program.QueueCompile()
	}

mainLoop:
	for {
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
					program.QueueCompile()
				} else {
					program.ClearBuild()
				}
			}
		case <-comp.ctx.Done():
			break mainLoop
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
		outputMap[buildPath] = program
	}

	return program
}

func (program *Program) QueueCompile() {
	if program.NeedRecompile() && program.compiling.CompareAndSwap(false, true) {
		inputPath := program.config.InputPath
		logs.Out("\n>>> Queued `%s`...\n", inputPath)
		go func() {
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
	}
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
