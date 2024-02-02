package proc

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"axlab.dev/bit/errs"
)

func Bootstrap() {
	exe := GetBootstrapExe()
	if exe == "" {
		return
	}

	rootDir := filepath.Join(ProjectDir(), "bootstrap")
	exeTime := errs.Handle(os.Stat(exe)).ModTime()
	queue := []string{rootDir}
	needReboot := false
	for len(queue) > 0 && !needReboot {
		entry := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if stat, err := os.Stat(entry); err == nil {
			if stat.ModTime().After(exeTime) {
				needReboot = true
			} else if stat.IsDir() {
				for _, it := range errs.Handle(os.ReadDir(entry)) {
					queue = append(queue, filepath.Join(entry, it.Name()))
				}
			}
		}
	}

	if needReboot {
		fmt.Printf("\nbootstrap: detected changes, rebuilding...\n")
		mainFile := filepath.Join(ProjectDir(), "bootstrap", "main.go")
		if Run("GO", "go", "build", "-o", exe, mainFile) {
			fmt.Printf("bootstrap: restarting...\n\n")
			files := make([]*os.File, 3)
			files[syscall.Stdin] = os.Stdin
			files[syscall.Stdout] = os.Stdout
			files[syscall.Stderr] = os.Stderr

			proc := errs.Handle(os.StartProcess(exe, os.Args, &os.ProcAttr{
				Dir:   WorkingDir(),
				Env:   os.Environ(),
				Files: files,
			}))

			status := errs.Handle(proc.Wait())
			os.Exit(status.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "bootstrap: rebuild failed\n")
		}
		fmt.Printf("\n")
	}
}

func GetBootstrapExe() string {
	exeFile := errs.Handle(os.Executable())

	if exeFile != "" {
		exeFile = errs.Handle(filepath.EvalSymlinks(exeFile))
	}

	if exeFile != "" {
		exeFile = filepath.Clean(exeFile)
	}

	if filepath.Dir(exeFile) != ProjectDir() {
		return "" // go run
	}

	return exeFile
}
