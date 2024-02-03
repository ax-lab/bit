package bit

import (
	"fmt"
	"path"
	"strings"
	"sync"
	"sync/atomic"
)

type ProgramConfig struct {
	InputPath     string
	BuildPath     string
	InputFullPath string
	BuildFullPath string
	LexerTemplate *Lexer
}

type Program struct {
	compiler *Compiler
	config   ProgramConfig

	lexer  *Lexer
	source *Source
	tokens []Token
	errors []error

	compiling  atomic.Bool
	buildMutex sync.Mutex
}

func NewProgram(compiler *Compiler, config ProgramConfig) *Program {
	return &Program{
		compiler: compiler,
		config:   config,
		lexer:    config.LexerTemplate.CopyOrDefault(),
	}
}

func (program *Program) Compile(source *Source) {
	program.source = source
	program.tokens = nil
	program.errors = nil

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
		tokenText.WriteString(fmt.Sprintf("\n\tat %s:%s", token.Span.Source().Name(), token.Pos.String()))
		if token.Span.Len() > 0 {
			pos := token.Pos
			pos.Advance(token.Span.Source().TabWidth(), token.Span.Text())
			tokenText.WriteString(fmt.Sprintf(" to %s", pos.String()))
		}
		tokenText.WriteString("\n\n")
	}
	program.writeOutput(tokenFile, tokenText.String())

	if errFile := baseName + ".errors.txt"; len(program.errors) > 0 {
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
	baseDir := program.source.Name()
	return path.Join(baseDir, name)
}
