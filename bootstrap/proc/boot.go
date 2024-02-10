package proc

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"axlab.dev/bit/common"
	"axlab.dev/bit/files"
)

var (
	interruptMutex sync.Mutex
	interruptChan  chan struct{}
)

func HandleInterrupt() chan struct{} {
	interruptMutex.Lock()
	defer interruptMutex.Unlock()
	if interruptChan == nil {
		interruptChan = make(chan struct{})
		inner := make(chan os.Signal, 1)
		signal.Notify(inner, os.Interrupt)
		go func() {
			for range inner {
				close(interruptChan)
			}
		}()
	}
	return interruptChan
}

func Bootstrap() {
	if rebuild, _ := NeedRebuild(); rebuild {
		if Rebuild() {
			common.Out("▸▸▸ bootstrap: restarting...\n\n")
			trap_interrupt := make(chan os.Signal, 1)
			signal.Notify(trap_interrupt, os.Interrupt)
			os.Exit(RunSelf(context.Background(), os.Args))
		}
	}
}

func NeedRebuild() (rebuild bool, newest time.Time) {
	exe := GetBootstrapExe(false)
	if exe == "" {
		return
	}

	rootDir := filepath.Join(files.ProjectDir(), "bootstrap")
	exeTime := common.Handle(os.Stat(exe)).ModTime()
	queue := []string{rootDir}
	for len(queue) > 0 {
		entry := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if stat, err := os.Stat(entry); err == nil {
			if mod := stat.ModTime(); mod.After(exeTime) {
				rebuild = true
				if newest.IsZero() || mod.After(newest) {
					newest = mod
				}
			} else if stat.IsDir() {
				for _, it := range common.Handle(os.ReadDir(entry)) {
					queue = append(queue, filepath.Join(entry, it.Name()))
				}
			}
		}
	}

	return
}

func Rebuild() bool {
	common.Out("\n▸▸▸ bootstrap: detected changes, rebuilding...\n")

	exe := GetBootstrapExe(true)
	mainFile := filepath.Join(files.ProjectDir(), "bootstrap", "main.go")

	success := Run("▸▸▸ GO", "go", "build", "-o", exe, mainFile)

	if !success {
		common.Out("\n")
		common.Err("▸▸▸ ERR: *******************************\n")
		common.Err("▸▸▸ ERR: *                             *\n")
		common.Err("▸▸▸ ERR: *   BOOTSTRAP BUILD FAILED!   *\n")
		common.Err("▸▸▸ ERR: *                             *\n")
		common.Err("▸▸▸ ERR: *******************************\n")
		common.Out("\n")
	}

	return success
}

func RunSelf(ctx context.Context, args []string) int {
	exe := GetBootstrapExe(true)

	fps := make([]*os.File, 3)
	fps[syscall.Stdin] = os.Stdin
	fps[syscall.Stdout] = os.Stdout
	fps[syscall.Stderr] = os.Stderr

	proc := common.Handle(os.StartProcess(exe, args, &os.ProcAttr{
		Dir:   files.WorkingDir(),
		Env:   os.Environ(),
		Files: fps,
	}))

	procFinished := make(chan struct{})
	shouldCancel := ctx.Done()
	if shouldCancel != nil {
		go func() {
			select {
			case <-shouldCancel:
				if runtime.GOOS == "windows" {
					proc.Kill()
				} else {
					proc.Signal(os.Interrupt)
					<-time.After(2 * time.Second)
					proc.Kill()
				}
			case <-procFinished:
			}
		}()
	}

	status := common.Handle(proc.Wait())
	close(procFinished)
	return status.ExitCode()
}

var bootstrapExe *string

func GetBootstrapExe(force bool) string {

	compute := func() string {
		exeFile := common.Handle(os.Executable())

		if exeFile != "" {
			exeFile = common.Handle(filepath.EvalSymlinks(exeFile))
		}

		if exeFile != "" {
			exeFile = filepath.Clean(exeFile)
		}

		if filepath.Dir(exeFile) != files.ProjectDir() {
			return "" // go run
		}

		return exeFile
	}

	if bootstrapExe == nil {
		exe := compute()
		bootstrapExe = &exe
	}

	out := *bootstrapExe
	if force && out == "" {
		panic("bootstrap: main executable not found")
	}
	return out
}

var bootstrapExeModTime *time.Time

func GetBootstrapExeModTime() time.Time {
	compute := func() (out time.Time) {
		exe := GetBootstrapExe(false)
		if exe == "" {
			return
		}

		if stat, err := os.Stat(exe); err == nil {
			out = stat.ModTime()
		}
		return
	}

	if bootstrapExeModTime == nil {
		value := compute()
		bootstrapExeModTime = &value
	}

	return *bootstrapExeModTime
}
