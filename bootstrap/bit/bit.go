package bit

import (
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"axlab.dev/bit/files"
	"axlab.dev/bit/logs"
	"axlab.dev/bit/proc"
)

type Compiler struct {
	inputDir files.Dir
	buildDir files.Dir

	programsMutex sync.Mutex
	programs      map[string]map[string]*Program
}

func NewCompiler(inputPath, buildPath string) *Compiler {
	return &Compiler{
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

	interrupt := proc.HandleInterrupt()
	inputEvents := watcher.Start(100 * time.Millisecond)

	logs.Sep()
	logs.Out(">>> Compiling from `%s` to `%s`...\n", input.Name(), comp.buildDir.Name())

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
		case <-interrupt:
			break mainLoop
		}
	}
}

type Program struct {
	compiler *Compiler

	inputPath     string
	buildPath     string
	inputFullPath string
	buildFullPath string

	compiling  atomic.Bool
	buildMutex sync.Mutex
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
		program = &Program{
			compiler:      comp,
			inputPath:     inputName,
			buildPath:     buildName,
			inputFullPath: inputPath,
			buildFullPath: buildPath,
		}
		outputMap[buildPath] = program
	}

	return program
}

func (program *Program) QueueCompile() {
	if program.compiling.CompareAndSwap(false, true) {
		logs.Sep()
		logs.Out("-> Queued `%s`...\n", program.inputPath)
		go func() {
			program.buildMutex.Lock()
			defer program.compiling.Store(false)
			defer program.buildMutex.Unlock()

			startTime := time.Now()
			logs.Break()
			logs.Out(".. Compiling `%s`...\n", program.inputPath)
			defer func() {
				duration := time.Since(startTime)
				logs.Break()
				logs.Out("== Finished `%s` in %s\n", program.inputPath, duration.String())
			}()

			compiler := program.compiler
			inputDir := compiler.inputDir
			buildDir := compiler.buildDir

			file := inputDir.ReadFile(program.inputPath)
			baseDir := file.Name()
			if file == nil {
				return
			}

			buildDir.Write(path.Join(baseDir, file.Name()+".src"), file.Text())
		}()
	}
}

func (program *Program) ClearBuild() {
	go func() {
		program.buildMutex.Lock()
		defer program.compiling.Store(false)
		defer program.buildMutex.Unlock()

		logs.Sep()
		logs.Out("-> Removing `%s`\n", program.inputPath)
		program.compiler.buildDir.RemoveAll(program.buildPath)
	}()
}
