package input

import (
	"fmt"
	"strings"
	"sync"
)

func Error(msg string, args ...any) ErrorWithLocation {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return ErrorWithLocation{
		text: msg,
	}
}

func ErrorAt(err error, src Source, pos ...int) error {
	return Error(err.Error()).At(src, pos...)
}

type ErrorWithLocation struct {
	loc  string
	text string
}

func (err ErrorWithLocation) String() string {
	return err.Error()
}

func (err ErrorWithLocation) Error() string {
	if err.loc == "" {
		return err.text
	}

	out := strings.Builder{}
	out.WriteString("in ")
	out.WriteString(err.loc)
	out.WriteString("\n\n")
	out.WriteString(Indent(err.text))

	return out.String()
}

func (err ErrorWithLocation) At(src Source, pos ...int) ErrorWithLocation {
	err.loc = Location(src.Name(), pos...)
	return err
}

func (err ErrorWithLocation) AtLocation(loc string) ErrorWithLocation {
	err.loc = loc
	return err
}

type ErrorList struct {
	mutex sync.RWMutex
	list  []error
}

func (errs *ErrorList) AddError(err error) {
	if err != nil {
		errs.mutex.Lock()
		defer errs.mutex.Unlock()
		errs.list = append(errs.list, err)
	}
}

func (errs *ErrorList) HasErrors() bool {
	errs.mutex.RLock()
	defer errs.mutex.RUnlock()
	return len(errs.list) > 0
}

func (errs *ErrorList) Errors() (out []error) {
	errs.mutex.RLock()
	defer errs.mutex.RUnlock()
	out = append(out, errs.list...)
	return
}

func (errs *ErrorList) ErrorText() string {
	errs.mutex.RLock()
	defer errs.mutex.RUnlock()

	out := strings.Builder{}
	for n, err := range errs.list {
		if n > 0 {
			out.WriteString("\n\n")
		}
		text := fmt.Sprintf("[%d] %s", n+1, err)
		out.WriteString(TrimSta(Indent(text)))
	}
	return out.String()
}
