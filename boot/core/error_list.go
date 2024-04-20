package core

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

func ErrorFromList(errs []error, msg string, args ...any) error {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	if len(errs) > 0 {
		out := strings.Builder{}
		out.WriteString(msg)
		out.WriteString(":")
		for n, err := range errs {
			out.WriteString(fmt.Sprintf("[%d] %v", n+1, err))
		}
		return errors.New(out.String())
	}

	return errors.New(msg)
}

type ErrorList struct {
	has  atomic.Bool
	sync sync.RWMutex
	list []error
	set  map[error]bool
}

func (errList *ErrorList) HasErrors() bool {
	return errList.has.Load()
}

func (errList *ErrorList) AddError(err error) {
	errList.AddErrors(err)
}

func (errList *ErrorList) AddErrors(errs ...error) {
	if len(errs) > 0 {
		errList.sync.Lock()
		defer errList.sync.Unlock()

		errList.has.Store(true)
		if errList.set == nil {
			errList.set = make(map[error]bool)
		}

		for _, err := range errs {
			if errList.set[err] {
				return
			}

			errList.set[err] = true
			errList.list = append(errList.list, err)
		}
	}
}

func (errList *ErrorList) Errors() (out []error) {
	if errList.HasErrors() {
		errList.sync.RLock()
		defer errList.sync.RUnlock()
		out = append(out, errList.list...)
	}
	return out
}
