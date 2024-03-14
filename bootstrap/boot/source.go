package boot

import (
	"os"
	"path/filepath"
	"sync"
)

type Source struct {
	st  *State
	err error

	Name string
	Text string
}

type sourceMap struct {
	mutex sync.Mutex
	files map[string]*Source
}

func (st *State) LoadSourceFile(file string) (*Source, error) {
	st.sourceMap.mutex.Lock()
	defer st.sourceMap.mutex.Unlock()

	fullPath, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	if src, ok := st.files[fullPath]; ok {
		if src.err != nil {
			return nil, src.err
		}
		return src, nil
	}

	src := &Source{
		st:   st,
		Name: file,
	}

	data, err := os.ReadFile(file)
	if err != nil {
		src.err = err
		return nil, err
	}

	src.Text = string(data)
	return src, nil
}
