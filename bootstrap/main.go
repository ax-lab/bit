package main

import (
	"os"
	"time"

	"axlab.dev/bit/files"
	"axlab.dev/bit/logs"
	"axlab.dev/bit/output"
	"axlab.dev/bit/proc"
	"axlab.dev/bit/text"
)

func main() {
	proc.Bootstrap()

	logs.Out("-> WorkDir: %s\n", proc.WorkingDir())
	logs.Out("-> Args:    %v\n", os.Args)

	logs.Out("-> Main:    %s\n", proc.FileName())
	logs.Out("-> Exe:     %s\n", proc.GetBootstrapExe())
	logs.Out("-> Project: %s\n", proc.ProjectDir())

	build := output.Open("./build")
	build.Write("src/main.c", text.Cleanup(`
		#include <stdio.h>

		int main() {
			printf("hello world\n");
			return 0;
		}
	`))

	if proc.Run("CC", "gcc", "./build/src/main.c", "-o", "./build/output.exe") {
		logs.Sep()
		exitCode := proc.Spawn("./build/output.exe")
		logs.Out("\nexited with %d\n", exitCode)
	} else {
		logs.Out("\nCompilation failed\n")
	}

	logs.Sep()
	logs.Out(">>> Listing %s\n", proc.WorkingDir())
	watcher := files.Watch(".", files.ListOptions{})
	logs.Sep()
	for _, it := range watcher.List() {
		logs.Break()
		logs.Out("- %s", it.String())
	}

	interrupt := proc.HandleInterrupt()

	logs.Sep()
	logs.Out(">>> Watching...\n")

	events := watcher.Start(100 * time.Millisecond)

outer:
	for {
		select {
		case list := <-events:
			for i, ev := range list {
				if i == 0 {
					logs.Sep()
				}
				logs.Break()
				logs.Out("%s\n", ev.String())
			}
		case <-interrupt:
			logs.Sep()
			logs.Out("Got interrupt")
			break outer
		}
	}

	logs.Sep()
}
