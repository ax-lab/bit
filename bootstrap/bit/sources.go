package bit

func (comp *Compiler) FlushSources() {
	comp.sourceFileMutex.Lock()
	defer comp.sourceFileMutex.Unlock()
	comp.sourceFileMap = nil
}

func (comp *Compiler) LoadSource(path string) (*Source, error) {
	fullPath, _, err := comp.inputDir.TryResolvePath(path)
	if err != nil {
		return nil, err
	}

	comp.sourceFileMutex.Lock()
	defer comp.sourceFileMutex.Unlock()

	if comp.sourceFileMap == nil {
		comp.sourceFileMap = make(map[string]*struct {
			src *Source
			err error
		})
	}

	entry := comp.sourceFileMap[fullPath]
	if entry == nil {
		file, err := comp.inputDir.TryReadFile(path)

		var src *Source
		if file != nil {
			src = file.CreateSource()
		}

		entry = &struct {
			src *Source
			err error
		}{src, err}
		comp.sourceFileMap[fullPath] = entry
	}

	return entry.src, entry.err
}
