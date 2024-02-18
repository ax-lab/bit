package bit

import (
	"fmt"
	"path"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"axlab.dev/bit/code"
	"axlab.dev/bit/proc"
)

const (
	debugCheckNodes = false
)

type ProgramConfig struct {
	InputPath     string
	BuildPath     string
	LexerTemplate *Lexer
	Globals       map[Key]Binding
}

type Program struct {
	Errors []error

	compiler *Compiler
	config   ProgramConfig

	lexer    *Lexer
	source   *Source
	tokens   []Token
	allNodes []*Node
	modules  []*Node
	mainNode *Node
	bindings *BindingMap
	scope    *code.Scope
	output   *code.Output

	coreInit   atomic.Bool
	compiling  atomic.Bool
	buildMutex sync.Mutex

	errMutex sync.Mutex
}

func NewProgram(compiler *Compiler, config ProgramConfig) *Program {
	return &Program{
		compiler: compiler,
		config:   config,
	}
}

func (program *Program) reset() {
	program.lexer = program.config.LexerTemplate.CopyOrDefault()
	program.source = nil
	program.tokens = nil
	program.Errors = nil
	program.allNodes = nil
	program.modules = nil
	program.mainNode = nil
	program.scope = code.NewScope(program)
	program.output = nil

	program.bindings = &BindingMap{
		program: program,
	}

	for key, binding := range program.config.Globals {
		program.bindings.BindGlobal(key, binding)
	}
}

func (program *Program) Valid() bool {
	return len(program.Errors) == 0
}

func (program *Program) BindNodes(key Key, nodes ...*Node) {
	program.bindings.AddNodes(key, nodes...)
}

func (program *Program) DeclareGlobal(key Key, binding Binding) {
	if program.config.Globals == nil {
		program.config.Globals = make(map[Key]Binding)
	}
	program.config.Globals[key] = binding
}

func (program *Program) NeedRecompile() bool {
	compiler := program.compiler
	input := compiler.inputDir.Stat(program.config.InputPath)
	if input == nil {
		return false
	}

	baseName := path.Base(program.config.InputPath)
	checkPath := program.outputPath(program.srcCopyName(baseName))
	output := compiler.buildDir.Stat(checkPath)
	if output == nil {
		return true
	}

	outputTime := output.ModTime()
	if outputTime.Before(input.ModTime()) {
		return true
	}

	if exeTime := proc.GetBootstrapExeModTime(); !exeTime.IsZero() && exeTime.After(outputTime) {
		return true
	}

	return false
}

func (program *Program) CompileSource(source *Source) {
	program.reset()
	program.source = source
	program.mainNode = program.loadSource(source)
	for program.bindings.StepNext() {
		if len(program.Errors) > 0 {
			break
		}
	}

	program.writeOutput("nodes.txt", program.dumpNodes(program.allNodes), true)

	var unresolved []*Node
	for _, it := range program.allNodes {
		if !it.IsDone() {
			unresolved = append(unresolved, it)
		}
	}

	const unresolvedFile = "errors-unresolved.txt"
	if cnt := len(unresolved); cnt > 0 {
		outputErrors := program.Valid()
		if outputErrors {
			program.HandleError(fmt.Errorf("there are %d unresolved nodes", cnt))
		}
		program.writeOutput(unresolvedFile, "# UNRESOLVED NODES\n\n"+program.dumpNodes(unresolved), true)

		if outputErrors {
			maxUnresolved := MaxErrorOutput - 1
			types, cnt := make(map[reflect.Type]bool), 0
			for _, it := range unresolved {
				if typ := reflect.TypeOf(it.Value()); !types[typ] {
					types[typ] = true
				} else if len(unresolved) > maxUnresolved {
					continue
				}

				cnt += 1
				if cnt > maxUnresolved {
					break
				}

				it.AddError("node is unresolved -- %s", it.Value().Repr(true))
			}
		}
	} else {
		program.removeOutput(unresolvedFile)
	}

	program.writeOutput("bindings.txt", program.bindings.Dump(), true)

	if debugCheckNodes {
		visited := make(map[*Node]bool)
		for _, it := range program.allNodes {
			checkNodes(it, visited)
		}
	}

	program.CompileOutput()
	program.writeOutput("code-output.txt", program.output.Repr(code.ReprDebug)+"\n", true)

	if errFile := "errors.txt"; len(program.Errors) > 0 {
		program.writeOutput(errFile, ErrorsToString(program.Errors, -1), true)
	} else {
		program.removeOutput(errFile)
	}

	for _, it := range program.modules {
		mod := it.Value().(Module)
		program.writeOutput("src/"+mod.Source.Name()+".dump.txt", it.Dump(true)+"\n", true)
	}
}

func (program *Program) CompileOutput() {
	node, valid := program.mainNode, program.Valid()
	if !valid {
		return
	} else if node == nil {
		panic("valid program must have a main node")
	}

	output := code.NewOutput(program.scope)
	outputCtx := output.NewContext()
	node.Output(outputCtx)
	if errs := output.Errors(); len(errs) == 0 {
		program.output = output
	} else {
		node.AddError("failed to output code for program")
		program.AddErrors(errs...)
	}
}

func (program *Program) generateCpp(outputDir, outputFile string) (mainPath string) {
	cppFiles := program.output.OutputCpp(outputFile)
	for name, text := range cppFiles {
		program.writeOutput(path.Join(outputDir, name), text, false)
	}

	mainFile := program.outputPath(path.Join(outputDir, outputFile))
	return program.compiler.buildDir.GetFullPath(mainFile)
}

func (program *Program) srcCopyName(baseName string) string {
	return "src/" + baseName + ".txt"
}

func (program *Program) loadSource(source *Source) *Node {
	program.bindings.InitSource(source)

	baseName := source.Name()
	program.writeOutput(program.srcCopyName(baseName), source.Text(), true)

	tokens, err := program.lexer.Tokenize(source)
	program.tokens = tokens
	program.HandleError(err)

	tokenFile := "tokens/" + baseName + ".txt"
	tokenText := strings.Builder{}
	for n, token := range program.tokens {
		if n > 0 {
			tokenText.WriteString("\n")
		}
		tokenText.WriteString(fmt.Sprintf("[%d of %d] %s", n+1, len(program.tokens), token.Type))
		if txt := token.Span.DisplayText(80); txt != "" {
			tokenText.WriteString(fmt.Sprintf(" = %s", txt))
		}
		tokenText.WriteString(fmt.Sprintf("\n\tat %s:%s", token.Span.Source().Name(), token.Span.Location().String()))
		if token.Span.Len() > 0 {
			pos := token.Span.Location()
			pos.Advance(token.Span.Source().TabWidth(), token.Span.Text())
			tokenText.WriteString(fmt.Sprintf(" to %s", pos.String()))
		}
		tokenText.WriteString("\n")
	}
	program.writeOutput(tokenFile, tokenText.String(), true)

	module := program.NewNode(Module{source}, source.Span())
	program.modules = append(program.modules, module)

	tokenNodes := make([]*Node, len(tokens))
	for i, it := range tokens {
		tokenNodes[i] = program.NewNode(it.Type, it.Span)
	}

	module.AddChildren(tokenNodes...)

	return module
}

func (program *Program) HandleError(err error) {
	if err != nil {
		program.AddErrors(err)
	}
}

func (program *Program) AddErrors(errs ...error) {
	program.errMutex.Lock()
	defer program.errMutex.Unlock()
	for _, it := range errs {
		if it != nil {
			program.Errors = append(program.Errors, it)
		}
	}
}

func (program *Program) dumpNodes(nodes []*Node) string {
	out := strings.Builder{}
	count := len(nodes)
	for n, it := range nodes {
		if n > 0 {
			out.WriteString("\n")
		}
		out.WriteString(fmt.Sprintf("[%03d / %03d] ", n+1, count))
		out.WriteString(fmt.Sprintf("%s #%d", it.value.Repr(true), it.id))

		if n := len(it.nodes); n > 0 {
			out.WriteString(fmt.Sprintf(" ==> [%d]{", n))
			for _, child := range it.nodes {
				out.WriteString(" ")
				out.WriteString(fmt.Sprintf("#%d", child.id))
			}
			out.WriteString(" }")
		}

		out.WriteString("\n\n\t@")
		out.WriteString(it.Span().String())
		if txt := it.Span().DisplayText(60); txt != "" {
			out.WriteString(" = ")
			out.WriteString(txt)
		}
		out.WriteString("\n")
	}
	return out.String()
}

func (program *Program) writeOutput(name, text string, withFooter bool) {
	var footer string
	if withFooter {
		footer = fmt.Sprintf("\n# GENERATED BY BUILD AT %s\n", time.Now().Format(time.RFC3339))
	}
	compiler := program.compiler
	path := program.outputPath(name)
	compiler.buildDir.Write(path, text+footer)
}

func (program *Program) removeOutput(name string) {
	compiler := program.compiler
	path := program.outputPath(name)
	compiler.buildDir.Remove(path)
}

func (program *Program) outputPath(name string) string {
	baseDir := program.config.BuildPath
	return path.Join(baseDir, name)
}

func checkNodes(node *Node, visited map[*Node]bool) {
	if visited[node] {
		return
	}

	if visited == nil {
		visited = make(map[*Node]bool)
	}
	visited[node] = true

	if node.parent != nil && node.parent.nodes[node.index] != node {
		panic(fmt.Sprintf("parent link for `%s` is invalid", node.String()))
	}

	for n, it := range node.nodes {
		if it.parent != node || it.index != n {
			panic(fmt.Sprintf("`%s` in parent `%s` is invalid", node.String(), it.String()))
		}
	}

	for _, it := range node.nodes {
		checkNodes(it, visited)
	}
}
