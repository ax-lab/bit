package proc

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"axlab.dev/bit/common"
	"axlab.dev/bit/logs"
)

// Spawn a new process "replacing" (not really) the current one.
//
// The new process shares the same environment and standard output streams
// as the current process.
//
// After the spawned process exits, the current process will exit with the
// same exit code.
func Replace(name string, args ...string) {
	os.Exit(Spawn(name, args...))
}

func Spawn(name string, args ...string) int {
	fullPath, err := exec.LookPath(name)
	if err != nil {
		if !errors.Is(err, exec.ErrDot) {
			logs.Check(err)
		}
	}

	files := make([]*os.File, 3)
	files[syscall.Stdin] = os.Stdin
	files[syscall.Stdout] = os.Stdout
	files[syscall.Stderr] = os.Stderr

	argv := []string{fullPath}
	argv = append(argv, args...)
	proc := logs.Handle(os.StartProcess(fullPath, argv, &os.ProcAttr{
		Dir:   ".",
		Env:   os.Environ(),
		Files: files,
	}))

	status := logs.Handle(proc.Wait())
	return status.ExitCode()
}

type CmdOutput struct {
	Success  bool
	StdOut   string
	StdErr   string
	Error    error
	ExitCode int
}

func Cmd(name string, args ...string) (out CmdOutput) {
	stdErr := strings.Builder{}
	stdOut := strings.Builder{}
	out.ExitCode, out.Error = Exec(name, args, func(output string, isError bool) {
		if isError {
			stdErr.WriteString(output)
		} else {
			stdOut.WriteString(output)
		}
	})

	out.StdOut = stdOut.String()
	out.StdErr = stdErr.String()
	out.Success = out.Error == nil && out.ExitCode == 0 && len(out.StdErr) == 0
	return out
}

// Run a command handling error output.
func Run(prefix, name string, args ...string) bool {
	stdErr, hasErr := strings.Builder{}, false
	status, err := Exec(name, args, func(output string, isError bool) {
		if isError {
			stdErr.WriteString(output)
			if !hasErr {
				logs.Out("\n")
			}
			hasErr = true
			if strings.ContainsAny(output, "\r\n") {
				lines := common.Lines(stdErr.String())
				for _, line := range lines[:len(lines)-1] {
					logs.Err("%s | %s\n", prefix, line)
				}
				stdErr.Reset()
				stdErr.WriteString(lines[len(lines)-1])
			}
		}
	})

	if stdErr.Len() > 0 {
		os.Stderr.WriteString(stdErr.String())
	}

	ok := err == nil && status == 0 && !hasErr

	if !ok {
		logs.Out("\n")
	}

	if err != nil {
		logs.Err("%s command error: %v\n", prefix, err)
	}

	if status != 0 {
		logs.Err("%s exited with %d\n", prefix, status)
	} else if hasErr {
		logs.Err("%s output errors", prefix)
	}

	return ok
}
