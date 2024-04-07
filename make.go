package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"axlab.dev/bit/boot/core"
)

const (
	DirBoot  = "boot"
	DirBuild = "build/bin"

	BootDirCmd = "cmd"
)

func main() {
	var (
		argVerbose = false
		argForce   = false
	)

	sta := time.Now()
	showDuration := func() {
		dur := time.Since(sta)
		fmt.Printf("=== Took %s\n\n", dur)
	}

	for _, arg := range os.Args[1:] {
		switch arg {
		case "-v", "--verbose":
			argVerbose = true
		case "-f", "--force":
			argForce = true
		}
	}

	project := FindProjectRoot()

	root := core.Check(core.FS(project))
	dirSrc := root.Get(DirBoot)
	dirBuild := root.Get(DirBuild)

	cmdSrc, cmdExe := FindCommands(dirSrc, dirBuild)

	var (
		srcTime time.Time
		srcPath string
	)
	srcFiles := core.CheckErrs(root.Glob("**.go"))
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
		build = build.SetDir(project)
		core.Handle(build.RunAndCheck())
	}
	showDuration()
}

func IsProjectRoot(path string) bool {
	make := filepath.Join(path, "make.go")
	boot := filepath.Join(path, "boot")
	return core.IsFile(make) && core.IsDir(boot)
}

func FindProjectRoot() string {
	cwd := core.Check(filepath.Abs("."))
	root := cwd
	for !IsProjectRoot(root) {
		next := filepath.Join(root, "..")
		if next == "" || next == root {
			core.Fatal(fmt.Errorf("could not find project path [cwd=%s]", cwd))
		}
		root = next
	}
	return root
}

func FindCommands(dirSrc, dirBuild core.File) (cmdSrc, cmdExe []core.File) {
	cmdList := core.Check(dirSrc.Get(BootDirCmd).List())
	for _, src := range cmdList {
		info := core.Check(src.Info())
		if !info.IsDir() {
			continue
		}

		exe := dirBuild.Get(src.Name() + ".exe")
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
