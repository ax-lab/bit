package lang

import (
	"fmt"
	"path"
	"strings"
	"unicode"

	"axlab.dev/bit/code"
	"axlab.dev/bit/core"
	"axlab.dev/bit/cpp"
)

func OutputC(compiler *core.Compiler, output *core.OutputSet) *core.CmdArgs {
	if compiler.HasErrors() {
		return nil
	}

	modules, moduleOutput := compiler.GetOutputCode()
	if len(modules) == 0 {
		return nil
	}

	cpp := cpp.Context{}
	for n, mod := range modules {
		output := moduleOutput[n]
		cppOutputModule(mod, &cpp, output)
		if compiler.ShouldStop() {
			break
		}
	}

	root := modules[0]
	name := path.Base(root.Name())
	main := name + ".c"
	cpp.GenerateOutput(main, output)

	output.WriteOutput()
	if compiler.HasErrors() {
		return nil
	}

	exeName := core.ExeName(name)
	exePath := core.Try(output.GetFullPath(exeName))
	srcPath := core.Try(output.GetFullPath(main))

	cppCmd := core.Cmd("gcc", "-Werror", srcPath, "-o", exePath)
	cppErr := cppCmd.Run()
	hasErr := false

	if errOutput := strings.TrimRightFunc(cppCmd.StdErr(), unicode.IsSpace); errOutput != "" {
		hasErr = true
		fmt.Fprintf(compiler.StdErr(), "CC error output:\n\n%s\n\n", core.Indent(errOutput, "\t|| "))
	}

	if cppErr != nil {
		cppErr = fmt.Errorf("CC failed to execute: %v", cppErr)
	} else if status := cppCmd.ExitCode(); status != 0 {
		cppErr = fmt.Errorf("CC exited with status %d", status)
	} else if hasErr {
		cppErr = fmt.Errorf("CC generated error output")
	}

	if cppErr != nil {
		root.Error(cppErr)
		return nil
	}

	exeCmd := core.Cmd(exePath)
	return exeCmd
}

func cppOutputModule(mod *core.Module, ctx *cpp.Context, expr []core.Expr) {
	fileName := mod.Name()
	initName := fileName
	initName = strings.ReplaceAll(initName, "/", "_")
	initName = strings.ReplaceAll(initName, "-", "_")
	initName = core.GetPlainIdentifier(initName)

	block := ctx.DeclareFunction(`void %s(void)`, initName)
	ctx.Main().WriteLine("%s();", initName)

	for _, it := range expr {
		cppOutputExpr(mod, block, it)
	}
}

func cppOutputExpr(mod *core.Module, block *cpp.Block, expr core.Expr) (out cpp.Var) {
	switch val := expr.(type) {
	case code.Seq:
		for _, it := range val.List() {
			out = cppOutputExpr(mod, block, it)
		}
		return out

	case code.Str:
		strVal := cpp.StringLiteral(val.Text())
		strVar := block.NewVar("str")
		block.Declare(`const char *%s = %s;`, strVar, strVal)
		return strVar

	case code.Print:

		var (
			argVar    []cpp.Var
			argVal    []core.Expr
			hasOutput cpp.Var
		)

		block.Context().IncludeSystem("stdio.h")
		for _, expr := range val.Args() {
			varName := cppOutputExpr(mod, block, expr)
			if varName != "" {
				argVar = append(argVar, varName)
				argVal = append(argVal, expr)
			}
		}

		for idx, varName := range argVar {
			if idx == 1 {
				hasOutput = block.NewVar("prn")
				block.Declare(`int %s = 0;`, hasOutput)
			}

			if idx > 0 {
				block.BlankLine()
				block.WriteLine(`if (%s) printf(" ");`, hasOutput)
			}
			cppOutputPrint(mod, block, varName, argVal[idx], hasOutput)
		}

		block.WriteLine(`printf("\n");`)
		return ""

	default:
		err := core.Errorf(val.Span(), "expression not supported in C output: %s", expr.String())
		mod.Error(err)
		return ""
	}
}

func cppOutputPrint(mod *core.Module, block *cpp.Block, argVar cpp.Var, argVal core.Expr, hasOutput cpp.Var) {
	switch val := argVal.(type) {
	case code.Str:
		if len(val.Text()) > 0 {
			if hasOutput != "" {
				block.WriteLine(`%s = 1;`, hasOutput)
			}
			block.WriteLine(`printf("%%s", %s);`, argVar)
		}

	default:
		err := core.Errorf(val.Span(), "expression not supported by C print: %s", val.String())
		mod.Error(err)
	}
}
