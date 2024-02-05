package bit

import (
	"fmt"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"axlab.dev/bit/proc"
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

	program.bindings = &BindingMap{}
	for key, binding := range program.config.Globals {
		program.bindings.BindGlobal(key, binding)
	}
	program.bindings.InitSource(source)

	baseName := source.Name()
	program.writeOutput(baseName+".src", source.Text())

	tokens, err := program.lexer.Tokenize(source)
	program.tokens = tokens
	program.HandleError(err)

	tokenFile := baseName + ".tokens.txt"
	tokenText := strings.Builder{}
	for n, token := range program.tokens {
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
		tokenText.WriteString("\n\n")
	}
	program.writeOutput(tokenFile, tokenText.String())

	if errFile := baseName + ".errors.txt"; len(program.errors) > 0 {
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
}

func (program *Program) HandleError(err error) {
	if err != nil {
		program.errMutex.Lock()
		defer program.errMutex.Unlock()
		program.errors = append(program.errors, err)
	}
}

func (program *Program) writeOutput(name, text string) {
	compiler := program.compiler
	path := program.outputPath(name)
	compiler.buildDir.Write(path, "# BUILD FILE\n\n"+text)
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
