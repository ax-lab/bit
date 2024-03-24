package bot

import "os"

func Run() {

	prog := ProgramNew()

	for _, it := range os.Args[1:] {
		prog.LoadFile(it)
	}

	prog.Run()

	if prog.HasErrors() {
		os.Exit(1)
	}

	// str := CodeStr("hello world!!!\n\n")
	// main := CodePrint{str}

	// rt := &Runtime{}
	// if _, err := rt.Eval(main); err != nil {
	// 	Fatal(err, "runtime error in eval")
	// }

	// const mainFile = "main.go"

	// cw := &CodeWriter{}
	// main.Output(cw)

	// out := NewOutput("run")
	// cw.Output(&out, mainFile)

	// if exitCode, err := out.Run(mainFile); err != nil {
	// 	Fatal(err, "failed to run `%s`", mainFile)
	// } else {
	// 	os.Exit(exitCode)
	// }
}
