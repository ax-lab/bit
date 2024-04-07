package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"axlab.dev/bit/boot/core"
)

const (
	dirBoot  = "boot"
	dirBuild = "build/bin"

	dirBootCmd = "cmd"
)

func main() {
	sta := time.Now()

	cwd := core.Check(filepath.Abs("."))
	root := cwd
	for !IsProjectRoot(root) {
		next := filepath.Join(root, "..")
		if next == "" || next == root {
			core.Fatal(fmt.Errorf("could not find project path from `%s`", cwd))
		}
		root = next
	}

	build := filepath.Join(root, dirBuild)
	core.Handle(os.MkdirAll(build, os.ModePerm))

	exeName := filepath.Base(os.Args[0])
	if ext := filepath.Ext(exeName); ext != "" {
		exeName = strings.TrimSuffix(exeName, ext)
	}

	srcDir := core.Check(core.FS(filepath.Join(root, dirBoot)))
	cmdDir := srcDir.Get(dirBootCmd + "/" + exeName)

	var cmdFile core.File
	if cmdDir.IsDir() {
		files := core.Check(cmdDir.List("*.go"))
		if len(files) == 1 {
			cmdFile = files[0]
		} else {
			for _, it := range files {
				if it.NameWithoutExt() == exeName {
					cmdFile = it
					break
				}
			}
		}
	}

	if !cmdFile.Valid() {
		core.Fatal(fmt.Errorf("could not find single source file for `%s`", exeName))
	}

	var buildTime time.Time

	exePath := filepath.Join(build, exeName+".exe")
	exeStat, exeStatErr := os.Stat(exePath)
	if exeStatErr != nil && !errors.Is(exeStatErr, fs.ErrNotExist) {
		core.Fatal(exeStatErr)
	} else if exeStatErr == nil {
		buildTime = exeStat.ModTime()
	}

	shouldRebuild := buildTime.IsZero()
	if shouldRebuild {
		fmt.Printf("(building `%s` from `%s`)\n\n", exeName, cmdFile.Path())
	} else {
		srcFiles := core.CheckErrs(srcDir.Glob("**.go"))

		for _, it := range srcFiles {
			info := core.Check(it.Info())
			if info.ModTime().After(buildTime) {
				fmt.Printf("(rebuilding `%s` -- %s modified)\n\n", exeName, it.Path())
				shouldRebuild = true
				break
			}
		}
	}

	if shouldRebuild {
		exe := core.Cmd("go", "build", "-o", exePath, path.Join(dirBoot, cmdFile.Path()))
		exe = exe.SetDir(root)
		core.Handle(exe.RunAndCheck())
	}

	dur := time.Since(sta)
	fmt.Println("Elapsed", dur)

	if exeName != "run" {
		cmd := core.Cmd(exePath, os.Args[1:]...).Pipe()
		core.Handle(cmd.Run())
		os.Exit(cmd.ExitCode())
	}
}

func IsProjectRoot(path string) bool {
	srcDir := filepath.Join(path, "src")
	gitDir := filepath.Join(path, ".git")
	return core.IsPath(gitDir) && core.IsDir(srcDir)
}
