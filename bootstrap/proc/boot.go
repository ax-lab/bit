package proc

import (
	"os"
	"path/filepath"
	"syscall"

	"axlab.dev/bit/logs"
)

func Bootstrap() {
	exe := GetBootstrapExe()
	if exe == "" {
		return
	}

	rootDir := filepath.Join(ProjectDir(), "bootstrap")
	exeTime := logs.Handle(os.Stat(exe)).ModTime()
	queue := []string{rootDir}
	needReboot := false
	for len(queue) > 0 && !needReboot {
		entry := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if stat, err := os.Stat(entry); err == nil {
			if stat.ModTime().After(exeTime) {
				needReboot = true
			} else if stat.IsDir() {
				for _, it := range logs.Handle(os.ReadDir(entry)) {
					queue = append(queue, filepath.Join(entry, it.Name()))
				}
			}
		}
	}

	if needReboot {
		logs.Sep()
		logs.Out("bootstrap: detected changes, rebuilding...\n")
		mainFile := filepath.Join(ProjectDir(), "bootstrap", "main.go")
		if Run("GO", "go", "build", "-o", exe, mainFile) {
			logs.Out("bootstrap: restarting...\n\n")
			files := make([]*os.File, 3)
			files[syscall.Stdin] = os.Stdin
			files[syscall.Stdout] = os.Stdout
			files[syscall.Stderr] = os.Stderr

			proc := logs.Handle(os.StartProcess(exe, os.Args, &os.ProcAttr{
				Dir:   WorkingDir(),
				Env:   os.Environ(),
				Files: files,
			}))

			status := logs.Handle(proc.Wait())
			os.Exit(status.ExitCode())
		} else {
			logs.Sep()
			logs.Err("bootstrap: rebuild failed\n")
		}
		logs.Sep()
	}
}

func GetBootstrapExe() string {
	exeFile := logs.Handle(os.Executable())

	if exeFile != "" {
		exeFile = logs.Handle(filepath.EvalSymlinks(exeFile))
	}

	if exeFile != "" {
		exeFile = filepath.Clean(exeFile)
	}

	if filepath.Dir(exeFile) != ProjectDir() {
		return "" // go run
	}

	return exeFile
}
