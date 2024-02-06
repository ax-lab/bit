package bit

import (
	"fmt"
	"path"
	"strings"
	"sync"
	"sync/atomic"

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
	compiler *Compiler
	config   ProgramConfig

	lexer    *Lexer
	source   *Source
	tokens   []Token
	errors   []error
	allNodes []*Node
	modules  []*Node
	bindings *BindingMap

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
	output := compiler.buildDir.Stat(program.outputPath(baseName + ".src"))
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

func (program *Program) Compile(source *Source) {
	program.lexer = program.config.LexerTemplate.CopyOrDefault()
	program.source = source
	program.tokens = nil
	program.errors = nil
	program.allNodes = nil
	program.modules = nil

	program.bindings = &BindingMap{
		program: program,
	}

	for key, binding := range program.config.Globals {
		program.bindings.BindGlobal(key, binding)
	}

	program.loadSource(source)
	for program.bindings.StepNext() {
		if len(program.errors) > 0 {
			break
		}
	}

	program.writeOutput("nodes.txt", program.dumpNodes(program.allNodes))

	var unresolved []*Node
	for _, it := range program.allNodes {
		if !it.Done() {
			unresolved = append(unresolved, it)
		}
	}

	const unresolvedFile = "errors-unresolved.txt"
	if cnt := len(unresolved); cnt > 0 {
		program.HandleError(fmt.Errorf("there are %d unresolved nodes", cnt))
		program.writeOutput(unresolvedFile, "# UNRESOLVED NODES\n\n"+program.dumpNodes(unresolved))
	} else {
		program.removeOutput(unresolvedFile)
	}

	program.writeOutput("bindings.txt", program.bindings.Dump())

	if debugCheckNodes {
		visited := make(map[*Node]bool)
		for _, it := range program.allNodes {
			checkNodes(it, visited)
		}
	}

	if errFile := "errors.txt"; len(program.errors) > 0 {
		SortErrors(program.errors)
		txt := strings.Builder{}
		for n, err := range program.errors {
			txt.WriteString(fmt.Sprintf("[%d of %d] ", n+1, len(program.errors)))
			txt.WriteString(err.Error())
			txt.WriteString("\n\n")
		}
		program.writeOutput(errFile, txt.String())
	} else {
		program.removeOutput(errFile)
	}

	for _, it := range program.modules {
		mod := it.Value().(Module)
		program.writeOutput("src/"+mod.Source.Name()+".dump.txt", it.Dump())
	}
}

func (program *Program) loadSource(source *Source) *Node {
	program.bindings.InitSource(source)

	baseName := source.Name()
	program.writeOutput("src/"+baseName, source.Text())

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
	program.writeOutput(tokenFile, tokenText.String())

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
		program.errMutex.Lock()
		defer program.errMutex.Unlock()
		program.errors = append(program.errors, err)
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
		out.WriteString(fmt.Sprintf("%s #%d", it.value.String(), it.id))

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

func (program *Program) writeOutput(name, text string) {
	compiler := program.compiler
	path := program.outputPath(name)
	compiler.buildDir.Write(path, text)
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
