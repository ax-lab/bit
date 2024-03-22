package bot

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

type CodeOutput struct {
	module    string
	files     map[string]string
	generated bool
	buildDir  string
}

func NewOutput(module string) CodeOutput {
	if !reModule.MatchString(module) {
		panic(fmt.Sprintf("invalid module name -- %s", module))
	}
	return CodeOutput{module: module, buildDir: "build"}
}

func (output *CodeOutput) AddFile(name, text string) {
	if output.files == nil {
		output.files = make(map[string]string)
	}
	output.files[name] = text
	output.generated = false
}

func (output *CodeOutput) Generate() {

	base := output.baseDir()

	sigText := output.Signature()
	sigFile := path.Join(base, "go.sig")

	if txt := strings.TrimSpace(ReadText(sigFile)); txt == sigText {
		output.generated = true
		return
	} else if reSignature.MatchString(txt) {
		if err := os.RemoveAll(base); err != nil {
			Fatal(err, "could not clean output directory")
		}
	} else if _, err := os.Stat(base); !os.IsNotExist(err) {
		Fatal(fmt.Errorf("output directory `%s` already exists", base), "cannot output code")
	}

	for name, text := range output.files {
		file := path.Join(base, name)
		WriteText(file, text)
	}

	WriteText(
		path.Join(base, "go.work"),
		"go "+goVersion+"\n\nuse (\n\t.\n)\n")

	WriteText(
		path.Join(base, "go.mod"),
		"module "+output.module+"\n\ngo "+goVersion+"\n")

	WriteText(sigFile, sigText)

	output.generated = true
}

func (output *CodeOutput) Run(file string) (exitCode int, err error) {
	if !output.generated {
		output.Generate()
	}

	exe := output.module + ".exe"
	cmd := exec.Command("go", "build", "-o", exe, file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = output.baseDir()
	if errBuild := cmd.Run(); errBuild != nil {
		err = fmt.Errorf("build failed: %v", errBuild)
		return
	}

	cmdExe := exec.Command(path.Join(cmd.Dir, exe))
	cmdExe.Stdout = os.Stdout
	cmdExe.Stderr = os.Stderr
	if errRun := cmdExe.Run(); errRun != nil {
		if exitErr, ok := (errRun).(*exec.ExitError); ok {
			exitCode = exitErr.ProcessState.ExitCode()
		} else {
			err = errRun
		}
	}
	return
}

func (output *CodeOutput) baseDir() string {
	return path.Join(output.buildDir, output.module)
}

func (output *CodeOutput) Signature() string {
	const hashSize = 256 / 8
	size := 0
	hash := sha256.New()
	for name, text := range output.files {
		size += len(text)
		_, err := hash.Write([]byte(name))
		if err == nil {
			_, err = hash.Write([]byte(text))
		}

		if err != nil {
			panic(fmt.Sprintf("hash returned error -- %v", err))
		}
	}

	var buffer [hashSize]byte
	sum := hash.Sum(buffer[:0])

	out := strings.Builder{}
	for _, b := range sum {
		out.WriteString(fmt.Sprintf("%02x", b))
	}
	out.WriteString(fmt.Sprintf(" (count=%d size=%d)", len(output.files), size))
	return out.String()
}

const (
	goVersion = "1.21.3"
)

var (
	reModule    = regexp.MustCompile(`^\w[_\w\d]*$`)
	reSignature = regexp.MustCompile(`^[0-9a-f]{64} \(count=\d+ size=\d+\)`)
)
