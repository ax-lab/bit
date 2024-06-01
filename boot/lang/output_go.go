package lang

import (
	"fmt"
	"path"
	"strings"
	"unicode"

	"axlab.dev/bit/code"
	"axlab.dev/bit/core"
	"axlab.dev/bit/golang"
)

func OutputGo(compiler *core.Compiler, output *core.OutputSet) *core.CmdArgs {
	if compiler.HasErrors() {
		return nil
	}

	modules, moduleOutput := compiler.GetOutputCode()
	if len(modules) == 0 {
		return nil
	}

	golang := golang.Context{}
	for n, mod := range modules {
		output := moduleOutput[n]
		goOutputModule(mod, &golang, output)
		if compiler.ShouldStop() {
			break
		}
	}

	root := modules[0]
	name := path.Base(root.Name())
	main := name + ".go"
	golang.GenerateOutput("bit/output/"+name, main, output)

	output.WriteOutput()
	if compiler.HasErrors() {
		return nil
	}

	exeName := core.ExeName(name)
	exePath := core.Try(output.GetFullPath(exeName))
	srcPath := core.Try(output.GetFullPath(main))

	goCmd := core.Cmd("go", "build", "-o", exePath, srcPath)
	goErr := goCmd.Run()
	hasErr := false

	if errOutput := strings.TrimRightFunc(goCmd.StdErr(), unicode.IsSpace); errOutput != "" {
		hasErr = true
		fmt.Fprintf(compiler.StdErr(), "GO error output:\n\n%s\n\n", core.Indent(errOutput, "\t|| "))
	}

	if goErr != nil {
		goErr = fmt.Errorf("GO failed to execute: %v", goErr)
	} else if status := goCmd.ExitCode(); status != 0 {
		goErr = fmt.Errorf("GO exited with status %d", status)
	} else if hasErr {
		goErr = fmt.Errorf("GO generated error output")
	}

	if goErr != nil {
		root.Error(goErr)
		return nil
	}

	exeCmd := core.Cmd(exePath)
	return exeCmd
}

func goOutputModule(mod *core.Module, ctx *golang.Context, expr []core.Expr) {
	fileName := mod.Name()
	initName := fileName
	initName = strings.ReplaceAll(initName, "/", "_")
	initName = strings.ReplaceAll(initName, "-", "_")
	initName = core.GetPlainIdentifier(initName)

	block := ctx.DeclareBlock(`func %s()`, initName)
	ctx.Main().WriteLine("%s()", initName)

	for _, it := range expr {
		goOutputExpr(mod, block, it)
	}
}

func goOutputExpr(mod *core.Module, block *golang.Block, expr core.Expr) (out golang.Var) {
	switch val := expr.(type) {
	case code.Seq:
		for _, it := range val.List() {
			out = goOutputExpr(mod, block, it)
		}
		return out

	case code.Str:
		strVal := golang.StringLiteral(val.Text())
		strVar := block.NewVar("str")
		block.Declare(`var %s = %s`, strVar, strVal)
		return strVar

	case code.Print:

		var (
			argVar    []golang.Var
			argVal    []core.Expr
			hasOutput golang.Var
		)

		block.Context().Import("fmt")
		for _, expr := range val.Args() {
			varName := goOutputExpr(mod, block, expr)
			if varName != "" {
				argVar = append(argVar, varName)
				argVal = append(argVal, expr)
			}
		}

		for idx, varName := range argVar {
			if idx == 1 {
				hasOutput = block.NewVar("prn")
				block.Declare(`var %s = false;`, hasOutput)
			}

			if idx > 0 {
				block.BlankLine()
				block.WriteLine(`if (%s) { fmt.Printf(" ") }`, hasOutput)
			}
			goOutputPrint(mod, block, varName, argVal[idx], hasOutput)
		}

		block.WriteLine(`fmt.Println()`)
		return ""

	default:
		err := core.Errorf(val.Span(), "expression not supported in Go output: %s", expr.String())
		mod.Error(err)
		return ""
	}
}

func goOutputPrint(mod *core.Module, block *golang.Block, argVar golang.Var, argVal core.Expr, hasOutput golang.Var) {
	switch val := argVal.(type) {
	case code.Str:
		if len(val.Text()) > 0 {
			if hasOutput != "" {
				block.WriteLine(`%s = true;`, hasOutput)
			}
			block.WriteLine(`fmt.Printf("%%s", %s);`, argVar)
		}

	default:
		err := core.Errorf(val.Span(), "expression not supported by Go print: %s", val.String())
		mod.Error(err)
	}
}
