package boot

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"axlab.dev/bit/input"
)

type ProgramError struct {
	src    input.Source
	line   int
	column int
	text   string
}

func Error(msg string, args ...any) ProgramError {
	err := ProgramError{
		text: msg,
	}
	if len(args) > 0 {
		err.text = fmt.Sprintf(err.text, args...)
	}
	return err
}

func ErrorAt(err error, src input.Source, pos ...int) error {
	return Error(err.Error()).At(src, pos...)
}

func (err ProgramError) String() string {
	return err.Error()
}

func (err ProgramError) Error() string {
	out := strings.Builder{}
	out.WriteString("in ")
	out.WriteString(err.src.Name())

	if err.line > 0 {
		out.WriteString(" @ L")
		out.WriteString(fmt.Sprintf("%03d", err.line))
		if err.column >= 0 {
			out.WriteString(":")
			out.WriteString(fmt.Sprintf("%02d", err.column))
		}
	}

	out.WriteString("\n\n\t")
	out.WriteString(err.text)
	return out.String()
}

func (err ProgramError) At(src input.Source, pos ...int) ProgramError {
	if !src.Valid() || len(pos) > 2 || (len(pos) > 0 && pos[0] <= 0) || (len(pos) > 1 && pos[1] <= 0) {
		panic("Error: invalid `at` position")
	}

	err.src = src
	if len(pos) > 0 {
		err.line = pos[0]
	}
	if len(pos) > 1 {
		err.column = pos[1]
	}
	return err
}

type errorList struct {
	mutex sync.Mutex
	list  []error
}

func (errs *errorList) AddError(err error) {
	if err != nil {
		errs.mutex.Lock()
		defer errs.mutex.Unlock()
		errs.list = append(errs.list, err)
	}
}

func (errs *errorList) Errors() (out []error) {
	errs.mutex.Lock()
	defer errs.mutex.Unlock()
	out = append(out, errs.list...)
	return
}

func (errs *errorList) CheckValid(stdErr io.Writer, prefix string) bool {
	list := errs.Errors()
	if len(list) == 0 {
		return true
	}

	fmt.Fprint(stdErr, prefix)
	for n, err := range list {
		if n > 0 {
			fmt.Fprintf(stdErr, "\n")
		}
		text := fmt.Sprintf("[%d] %s\n", n+1, err)
		fmt.Fprint(stdErr, StrIndent(text))
	}
	return false
}
