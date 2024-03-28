package bot

import "os"

func Run() {

	prog := ProgramNew()

	for _, it := range os.Args[1:] {
		prog.LoadFile(it)
	}

	prog.Eval()

	const mainFileName = "main.go"
	goProgram, mainFile := GoProgramNew("axlab.dev/output", mainFileName)
	prog.GoOutput(goProgram, mainFile)

	if !prog.HasErrors() {
		output := NewOutput(goProgram.Module())
		goProgram.OutputTo(&output)
		if exitCode, err := output.Run(mainFileName); err != nil {
			Fatal(err, "failed to run `%s`", mainFile)
		} else {
			os.Exit(exitCode)
		}
	}

	prog.Run()
	if prog.HasErrors() {
		os.Exit(1)
	}
}
