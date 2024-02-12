package proc

import (
	"io"
	"os"
	"os/exec"
	"sync"

	"axlab.dev/bit/common"
)

func ExecInDir(prefix, dir string, callback func() bool) bool {
	var success bool
	cwd := common.Try(os.Getwd())
	common.Check(os.Chdir(dir))
	success = callback()
	common.Check(os.Chdir(cwd))
	return success
}

func ExecWithStream(name string, args []string, stdOut io.Writer, stdErr io.Writer) (int, error) {
	if stdOut == nil {
		stdOut = os.Stdout
	}
	if stdErr == nil {
		stdErr = os.Stderr
	}
	return Exec(name, args, func(output string, isError bool) {
		if isError {
			io.WriteString(stdErr, output)
		} else {
			io.WriteString(stdOut, output)
		}
	})
}

// Exec a process using a callback to process output.
func Exec(name string, args []string, callback func(output string, isError bool)) (int, error) {
	cmd := exec.Command(name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return -1, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return -1, err
	}

	if err = cmd.Start(); err != nil {
		return -1, err
	}

	wg := sync.WaitGroup{}

	consume := func(output io.Reader, isError bool) {
		buffer := [1024]byte{}
		defer wg.Done()
		for {
			n, err := output.Read(buffer[:])
			if n > 0 {
				callback(string(buffer[:n]), isError)
			}
			if err == io.EOF {
				break
			}
		}
	}

	wg.Add(2)
	go consume(stdout, false)
	go consume(stderr, true)

	status, err := cmd.Process.Wait()
	wg.Wait()
	return status.ExitCode(), err
}
