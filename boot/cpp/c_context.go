package cpp

import (
	"fmt"
	"sync"

	"axlab.dev/bit/core"
)

type Context struct {
	sync sync.Mutex

	includeLocalSet  map[string]bool
	includeLocalList []string

	includeSystemSet  map[string]bool
	includeSystemList []string

	declHead []string
	declBody []*Block

	mainDecl Block
	mainBody Block
}

func (cpp *Context) IncludeLocal(includes ...string) {
	cpp.sync.Lock()
	defer cpp.sync.Unlock()
	if cpp.includeLocalSet == nil {
		cpp.includeLocalSet = make(map[string]bool)
	}
	for _, it := range includes {
		if !cpp.includeLocalSet[it] {
			cpp.includeLocalSet[it] = true
			cpp.includeLocalList = append(cpp.includeLocalList, it)
		}
	}
}

func (cpp *Context) IncludeSystem(includes ...string) {
	cpp.sync.Lock()
	defer cpp.sync.Unlock()
	if cpp.includeSystemSet == nil {
		cpp.includeSystemSet = make(map[string]bool)
	}
	for _, it := range includes {
		if !cpp.includeSystemSet[it] {
			cpp.includeSystemSet[it] = true
			cpp.includeSystemList = append(cpp.includeSystemList, it)
		}
	}
}

func (cpp *Context) Main() *Block {
	cpp.initBlocks()
	return &cpp.mainBody
}

func (cpp *Context) DeclareFunction(header string, args ...any) *Block {
	if len(args) > 0 {
		header = fmt.Sprintf(header, args...)
	}

	cpp.mainDecl.BlankLine()
	cpp.mainDecl.WriteLine("%s;", header)

	body := &Block{context: cpp}
	cpp.declHead = append(cpp.declHead, header)
	cpp.declBody = append(cpp.declBody, body)

	return body
}

func (cpp *Context) GenerateOutput(mainFile string, output *core.OutputSet) {
	file := core.CodeText{}

	for _, it := range cpp.includeSystemList {
		file.WriteLine(`#include <%s>`, it)
	}
	if len(cpp.includeSystemList) > 0 {
		file.BlankLine()
	}

	for _, it := range cpp.includeLocalList {
		file.WriteLine(`#include "%s"`, it)
	}
	if len(cpp.includeLocalList) > 0 {
		file.BlankLine()
	}

	file.WriteLine(cpp.mainDecl.String())

	for idx, header := range cpp.declHead {
		body := cpp.declBody[idx]
		file.BlankLine()
		file.WriteLine("%s {", header)
		file.Indent()
		file.WriteLine(body.String())
		file.Dedent()
		file.WriteLine("}")
	}

	file.BlankLine()
	file.WriteLine(`int main() {`)
	file.Indent()
	file.WriteLine(cpp.mainBody.String())
	file.Dedent()
	file.WriteLine(`}`)

	text := file.String()
	output.Add(mainFile, text)
}

func (cpp *Context) initBlocks() {
	cpp.mainDecl.context = cpp
	cpp.mainBody.context = cpp
}
