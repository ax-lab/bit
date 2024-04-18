package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"axlab.dev/bit/boot/core"
)

const (
	DirBoot  = "boot"
	DirBuild = "build/bin"

	BootDirCmd = "cmd"
)

var (
	projectDir = core.ProjectRoot()
)

func main() {

	exe := core.Check(os.Executable())
	exe = core.ExeName(filepath.Base(exe))

	src := core.ExeName(CurFile())
	isSelf := src == exe
	build(isSelf)

	if !isSelf {
		exePath := filepath.Join(projectDir, DirBuild, exe)
		cmd := core.Cmd(exePath, os.Args[1:]...).Pipe()
		core.Handle(cmd.Run())
		os.Exit(cmd.ExitCode())
	}
}

func build(isSelf bool) {
	var (
		argVerbose = false
		argForce   = false
	)

	sta := time.Now()
	showDuration := func() {
		dur := time.Since(sta)
		fmt.Printf("=== Took %s\n\n", dur)
	}

	if isSelf {
		for _, arg := range os.Args[1:] {
			switch arg {
			case "-v", "--verbose":
				argVerbose = true
			case "-f", "--force":
				argForce = true
			}
		}
	}

	RebuildSelf()

	root := core.Check(core.FS(projectDir))
	dirSrc := root.Get(DirBoot)
	dirBuild := root.Get(DirBuild)

	cmdSrc, cmdExe := FindCommands(dirSrc, dirBuild)

	var (
		srcTime time.Time
		srcPath string
	)
	srcFiles := core.CheckErrs(root.Glob("boot/**.go"))
	for _, src := range srcFiles {
		modTime := core.Check(src.Info()).ModTime()
		if srcTime.IsZero() || modTime.After(srcTime) {
			srcTime = modTime
			srcPath = src.Path()
		}
	}

	if srcPath == "" {
		core.Fatal(fmt.Errorf("no source files found to build"))
	}

	var buildIndex []int
	for n, exe := range cmdExe {
		var rebuild bool
		if argForce {
			rebuild = true
		} else if exe.Exists() {
			info := core.Check(exe.Info())
			if info.ModTime().Before(srcTime) {
				rebuild = true
			}
		} else {
			rebuild = true
		}

		if rebuild {
			buildIndex = append(buildIndex, n)
		}
	}

	if len(buildIndex) == 0 {
		if argVerbose {
			fmt.Printf(">>> Up to date, nothing to build...\n")
			showDuration()
		}
		return
	}

	if argForce {
		srcPath += ", forced"
	}
	fmt.Printf(">>> Modified, rebuilding... (%s)\n", srcPath)
	for _, idx := range buildIndex {
		exe := cmdExe[idx]
		src := cmdSrc[idx]
		fmt.Printf("... Building `%s` to `%s`\n", src.Path(), exe.Path())

		build := core.Cmd("go", "build", "-o", exe.FilePath(), src.FilePath())
		build = build.SetDir(projectDir)
		core.Handle(build.RunAndCheck())
	}
	showDuration()
}

func CurFile() string {
	_, file, _, _ := runtime.Caller(0)
	file = filepath.Base(file)
	return file
}

func RebuildSelf() {
	file := CurFile()
	if file != "" {
		glob := path.Join(projectDir, "**.go")
		src := filepath.Join(projectDir, file)
		exeName := core.ExeName(file)
		exe := filepath.Join(projectDir, exeName)
		if core.Check(core.NeedRebuild(exe, glob)) {
			fmt.Printf(">>> Rebuilding %s...\n", exeName)
			build := core.Cmd("go", "build", "-o", exe, src)
			build = build.SetDir(projectDir)
			core.Handle(build.RunAndCheck())
		}
	}
}

func FindCommands(dirSrc, dirBuild core.File) (cmdSrc, cmdExe []core.File) {
	cmdList := core.Check(dirSrc.Get(BootDirCmd).List())
	for _, src := range cmdList {
		info := core.Check(src.Info())
		if !info.IsDir() {
			continue
		}

		exe := dirBuild.Get(core.ExeName(src.Name()))
		exeStat, exeErr := exe.Info()
		if exeErr != nil {
			if !errors.Is(exeErr, fs.ErrNotExist) {
				core.Handle(exeErr)
			}
		} else if exeStat.IsDir() {
			core.Fatal(fmt.Errorf("output exe `%s` is a directory", exe.Path()))
		}

		cmdSrc = append(cmdSrc, src)
		cmdExe = append(cmdExe, exe)
	}

	if len(cmdSrc) == 0 {
		core.Fatal(fmt.Errorf("no command executables to build"))
	}

	return
}