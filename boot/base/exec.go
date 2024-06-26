package base

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"
)

func ExeName(name string) string {
	exeName := name
	if ext := path.Ext(exeName); ext != "" && len(ext) < len(exeName) {
		exeName = exeName[:len(exeName)-len(ext)]
	}

	if runtime.GOOS == "windows" {
		return exeName + ".exe"
	}
	return exeName
}

type CmdArgs struct {
	// config
	name      string
	args      []string
	workDir   string
	pipeStdio bool

	// status
	done    sync.WaitGroup
	started atomic.Bool
	exeDone atomic.Bool

	// result
	err       error
	stdOut    []byte
	stdErr    []byte
	stdOutErr error
	stdErrErr error
	state     *os.ProcessState
}

func Cmd(name string, args ...string) *CmdArgs {
	cmd := &CmdArgs{
		name: name,
		args: args,
	}
	cmd.done.Add(1) // wait for cmd.Start()
	return cmd
}

func (cmd *CmdArgs) WorkDir(dir string) *CmdArgs {
	cmd.workDir = dir
	return cmd
}

func (cmd *CmdArgs) PipeStdio() *CmdArgs {
	cmd.pipeStdio = true
	return cmd
}

func (cmd *CmdArgs) String() string {
	out := strings.Builder{}
	out.WriteString("Cmd(")
	out.WriteString(cmd.name)
	for n, it := range cmd.args {
		if n == 0 {
			out.WriteString(", args: ")
		} else {
			out.WriteString(" ")
		}
		out.WriteString(fmt.Sprintf("%#v", it))
	}

	if cmd.workDir != "" {
		out.WriteString(", dir=")
		out.WriteString(cmd.workDir)
	}

	if cmd.pipeStdio {
		out.WriteString(", pipe=stdio")
	}

	out.WriteString(")")
	return out.String()
}

func (cmd *CmdArgs) Done() bool {
	return cmd.exeDone.Load()
}

func (cmd *CmdArgs) Run() error {
	cmd.Start()
	cmd.Wait()
	return cmd.Error()
}

func (cmd *CmdArgs) Wait() {
	cmd.done.Wait()
}

func (cmd *CmdArgs) Error() error {
	return cmd.err
}

func (cmd *CmdArgs) ExitCode() int {
	if cmd.state != nil {
		return cmd.state.ExitCode()
	}
	return 0
}

func (cmd *CmdArgs) StdOut() string {
	return string(cmd.stdOut)
}

func (cmd *CmdArgs) StdErr() string {
	return string(cmd.stdErr)
}

func (cmd *CmdArgs) RunAndCheck() error {
	err := cmd.Run()
	if err == nil {
		hasErrors := false
		if stdErr := strings.TrimRightFunc(cmd.StdErr(), unicode.IsSpace); stdErr != "" {
			fmt.Fprintf(os.Stderr, "Error output:\n\n%s\n\n", Indent(stdErr, Prefix("    ")))
			hasErrors = true
		}
		if code := cmd.ExitCode(); code != 0 {
			err = fmt.Errorf("command `%s` failed with status %d", cmd.name, code)
		} else if hasErrors {
			err = fmt.Errorf("command `%s` generated errors", cmd.name)
		}
	}
	return err
}

func (cmd *CmdArgs) Start() {
	if !cmd.started.CompareAndSwap(false, true) {
		panic("Run: command has already started")
	}

	complete := func(err error) {
		if err != nil && cmd.err == nil {
			err = fmt.Errorf("cmd `%s`: %v", cmd.name, err)
			cmd.err = err
		}
		if cmd.exeDone.CompareAndSwap(false, true) {
			cmd.done.Done()
		}
	}

	exe := exec.Command(cmd.name, cmd.args...)
	exe.Dir = cmd.workDir

	start := func() {
		if err := exe.Start(); err != nil {
			complete(err)
			return
		}
	}

	if cmd.pipeStdio {
		exe.Stderr = os.Stderr
		exe.Stdout = os.Stdout
		exe.Stdin = os.Stdin
		start()
	} else {

		stdout, err := exe.StdoutPipe()
		if err != nil {
			complete(fmt.Errorf("piping stdout: %v", err))
			return
		}

		stderr, err := exe.StderrPipe()
		if err != nil {
			complete(fmt.Errorf("piping stderr: %v", err))
			return
		}

		start()

		consume := func(output io.Reader, isError bool) {
			buffer := [1024]byte{}
			defer cmd.done.Done()
			for {
				n, err := output.Read(buffer[:])
				cmd.processOutput(buffer[:n], err, isError)
				if err != nil {
					break
				}
			}
		}

		cmd.done.Add(2)
		go consume(stdout, false)
		go consume(stderr, true)
	}

	go func() {
		defer complete(nil)
		if exe.Process == nil {
			return
		}
		state, err := exe.Process.Wait()
		cmd.state = state
		complete(err)
	}()
}

func (cmd *CmdArgs) processOutput(buffer []byte, err error, isError bool) {
	if isError {
		cmd.stdErr = append(cmd.stdErr, buffer...)
		cmd.stdErrErr = err
	} else {
		cmd.stdOut = append(cmd.stdOut, buffer...)
		cmd.stdOutErr = err
	}
}
