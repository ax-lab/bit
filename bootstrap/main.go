package main

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"axlab.dev/bit/bit"
	"axlab.dev/bit/common"
	"axlab.dev/bit/proc"
)

const (
	SampleC = false

	FlagBoot   = "--boot"
	FlagBooted = "--bootstrapped"
	FlagOnce   = "--once"
)

func main() {

	args := os.Args

	var boot, booted, once bool
	if len(args) > 1 {
		switch args[1] {
		case FlagBoot:
			boot = true
		case FlagBooted:
			booted = true
		case FlagOnce:
			once = true
		}
	}

	if !booted {
		proc.Bootstrap()
	}

	if boot {
		newArgs := append([]string{}, args...)
		newArgs[1] = FlagBooted
		BootWatcher(newArgs)
	} else {

		ctx, cancel := context.WithCancel(context.Background())

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)

		go func() {
			<-interrupt
			cancel()
		}()

		compiler := bit.NewCompiler(ctx, "sample", "build")
		inputDir := compiler.InputDir()
		buildDir := compiler.BuildDir()

		common.Out("○○○ Input: %s\n", inputDir.FullPath())
		common.Out("○○○ Build: %s\n", buildDir.FullPath())

		compiler.Watch(once)
		if once {
			common.Out("\n")
		}

		if SampleC {
			main := buildDir.Write("src/main.c", common.CleanupText(`
			#include <stdio.h>

			int main() {
				printf("hello world\n");
				return 42;
			}
		`))

			output := buildDir.GetFullPath("output.exe")
			if proc.Run("CC", "gcc", main.FullPath(), "-o", output) {
				common.Out("\n")
				if exitCode := proc.Spawn(output); exitCode != 0 {
					common.Out("\n(exited with %d)\n", exitCode)
				} else {
					common.Out("\n")
				}
			} else {
				common.Out("\nCompilation failed\n")
			}
		}
	}
}

func BootWatcher(newArgs []string) {
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
