package base

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type ErrorSet struct {
	sync sync.RWMutex
	has  map[error]bool
	list []error
}

func (set *ErrorSet) Add(errs ...error) {
	if len(errs) == 0 {
		return
	}

	set.sync.Lock()
	defer set.sync.Unlock()
	set.addList(errs)
}

func (set *ErrorSet) addList(errs []error) {
	for _, err := range errs {
		set.addError(err)
	}
}

func (set *ErrorSet) addError(err error) {
	if err == nil || set.has[err] {
		return
	}

	if set.has == nil {
		set.has = make(map[error]bool)
	}
	set.has[err] = true

	if subset, isSet := err.(*ErrorSet); isSet {
		set.addList(subset.list)
	} else {
		set.list = append(set.list, err)
	}
}

func (set *ErrorSet) Len() int {
	set.sync.RLock()
	defer set.sync.RUnlock()
	return len(set.list)
}

func (set *ErrorSet) Errors() (out []error) {
	set.sync.RLock()
	defer set.sync.RUnlock()
	out = append(out, set.list...)
	return out
}

func (set *ErrorSet) Is(target error) bool {
	set.sync.RLock()
	defer set.sync.RUnlock()
	for _, err := range set.list {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func (set *ErrorSet) Unwrap() []error {
	set.sync.RLock()
	defer set.sync.RUnlock()
	return set.list
}

func (set *ErrorSet) Error() string {
	return set.String()
}

func (set *ErrorSet) String() string {
	set.sync.RLock()
	defer set.sync.RUnlock()

	if len(set.list) == 0 {
		return ""
	}

	out := strings.Builder{}
	out.WriteString(fmt.Sprintf("there were %d errors:", len(set.list)))

	for n, it := range set.list {
		out.WriteString(fmt.Sprintf("\n[%d] %s", n+1, it.Error()))
	}

	return out.String()
}

func Error(msg string, args ...any) (err error) {
	if len(args) == 0 {
		err = errors.New(msg)
	} else {
		err = fmt.Errorf(msg, args...)
	}
	return err
}

func Errors(args ...error) (err error) {
	set := ErrorSet{}
	set.Add(args...)

	list := set.Unwrap()
	switch len(list) {
	case 0:
		return nil
	case 1:
		return list[0]
	default:
		return &set
	}
}
