package main

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"axlab.dev/bit/bit"
	"axlab.dev/bit/logs"
	"axlab.dev/bit/proc"
	"axlab.dev/bit/text"
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

		logs.Out("○○○ Input: %s\n", inputDir.FullPath())
		logs.Out("○○○ Build: %s\n", buildDir.FullPath())

		compiler.Watch(once)
		if once {
			logs.Out("\n")
		}

		if SampleC {
			main := buildDir.Write("src/main.c", text.Cleanup(`
			#include <stdio.h>

			int main() {
				printf("hello world\n");
				return 42;
			}
		`))

			output := buildDir.GetFullPath("output.exe")
			if proc.Run("CC", "gcc", main.FullPath(), "-o", output) {
				logs.Out("\n")
				if exitCode := proc.Spawn(output); exitCode != 0 {
					logs.Out("\n(exited with %d)\n", exitCode)
				} else {
					logs.Out("\n")
				}
			} else {
				logs.Out("\nCompilation failed\n")
			}
		}
	}
}

func BootWatcher(newArgs []string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	logs.Out("▸▸▸ bootstrap: watching for changes...\n")

	runSelf := func() (cancel func()) {
		ctx, cancelCtx := context.WithCancel(context.Background())

		var cancelled atomic.Bool
		done := make(chan struct{})
		go func() {
			defer close(done)
			logs.Out("▸▸▸ STARTING...\n\n")
			exitCode := proc.RunSelf(ctx, newArgs)

			mode := ""
			if cancelled.Load() {
				mode = " (cancelled)"
			}

			logs.Out("\n▸▸▸ EXIT STATUS %d%s\n\n", exitCode, mode)
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
			logs.Out("\n▸▸▸ INTERRUPTED\n\n")
			cancel()
			break mainLoop
		case <-time.After(500 * time.Millisecond):
			if rebuild, timestamp := proc.NeedRebuild(); rebuild && timestamp.After(lastCheck) {
				lastCheck = timestamp
				if proc.Rebuild() {
					logs.Out("▸▸▸ bootstrap: restarting child...\n\n")
					cancel()
					cancel = runSelf()
				}
			}
		}
	}
}
