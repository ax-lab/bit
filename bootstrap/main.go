package main

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"axlab.dev/bit/bit"
	"axlab.dev/bit/bit_core"
	"axlab.dev/bit/common"
	"axlab.dev/bit/proc"
)

const (
	FlagBoot   = "--boot"
	FlagBooted = "--bootstrapped"
	FlagWatch  = "--watch"
	FlagCpp    = "--cpp"

	BuildDir = "build"
)

func main() {

	args := os.Args

	var boot, booted, watch, cpp bool
	var files []string
	if len(args) > 1 {
		skip := true
		switch args[1] {
		case FlagBoot:
			boot = true
		case FlagBooted:
			booted = true
		case FlagWatch:
			watch = true
		case FlagCpp:
			cpp = true
		default:
			skip = false
		}
		if skip {
			files = args[2:]
		} else {
			files = args[1:]
		}
	}

	// recompile the binary before running, unless we are running from the watcher
	if !booted {
		proc.Bootstrap()
	}

	if boot {
		newArgs := append([]string{}, args...)
		newArgs[1] = FlagBooted
		BootstrapWatcher(newArgs)
	} else if booted || watch {
		WatchAndCompile()
	} else {
		if len(files) == 0 {
			common.Out("\nNo files giving, exiting\n\n")
			return
		}
		compiler := bit.NewCompiler(context.Background(), ".", BuildDir+"/run")
		compiler.SetCore(bit_core.InitCompiler)
		for _, it := range files {
			res := compiler.Run(it, bit.RunOptions{Cpp: cpp})
			if len(files) > 1 {
				common.Out("\n>>> %s <<<\n", it)
			}

			hasOutput := false
			if res.Err != nil {
				hasOutput = true
				common.Err("\nError: %v\n", res.Err)
			}

			if len(res.Log) > 0 {
				hasOutput = true
				errText := common.Indented(common.ErrorsToString(res.Log, common.MaxErrorOutput))
				common.Err("\n\t>>> Error Log <<<\n\n\t" + errText + "\n")
			}

			if hasOutput {
				common.Out("\n")
			}

			common.Out("\n%s\n\n", res.Repr())
		}
	}
}

func WatchAndCompile() {
	ctx, cancel := context.WithCancel(context.Background())

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		<-interrupt
		cancel()
	}()

	compiler := bit.NewCompiler(ctx, "sample", BuildDir+"/sample")
	compiler.SetCore(bit_core.InitCompiler)
	inputDir := compiler.InputDir()
	buildDir := compiler.BuildDir()

	common.Out("○○○ Input: %s\n", inputDir.FullPath())
	common.Out("○○○ Build: %s\n", buildDir.FullPath())

	compiler.Watch()
	common.Out("\n")
}

func BootstrapWatcher(newArgs []string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	common.Out("▸▸▸ bootstrap: watching for changes...\n")

	runSelf := func() (cancel func()) {
		ctx, cancelCtx := context.WithCancel(context.Background())

		var cancelled atomic.Bool
		done := make(chan struct{})
		go func() {
			defer close(done)
			common.Out("▸▸▸ STARTING...\n\n")
			exitCode := proc.RunSelf(ctx, newArgs)

			mode := ""
			if cancelled.Load() {
				mode = " (cancelled)"
			}

			common.Out("\n▸▸▸ EXIT STATUS %d%s\n\n", exitCode, mode)
			cancelCtx()
		}()

		cancel = func() {
			cancelled.Store(true)
			cancelCtx()
			<-done
		}
		return cancel
	}

	cancel := runSelf()

	lastCheck := time.Time{}

mainLoop:
	for {
		select {
		case <-interrupt:
			common.Out("\n▸▸▸ INTERRUPTED\n\n")
			cancel()
			break mainLoop
		case <-time.After(500 * time.Millisecond):
			if rebuild, timestamp := proc.NeedRebuild(); rebuild && timestamp.After(lastCheck) {
				lastCheck = timestamp
				if proc.Rebuild() {
					common.Out("▸▸▸ bootstrap: restarting child...\n\n")
					cancel()
					cancel = runSelf()
				}
			}
		}
	}
}
