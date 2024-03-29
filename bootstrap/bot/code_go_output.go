package bot

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

func (program *GoProgram) OutputTo(output *CodeOutput) {
	for name, file := range program.files {
		text := file.OutputText()
		output.AddFile(name, text)
	}
}

func (file *GoFile) OutputText() string {
	text := strings.Builder{}

	text.WriteString(fmt.Sprintf("package %s\n", file.module))

	if len(file.imports) > 0 {
		var imports []string
		for name := range file.imports {
			imports = append(imports, name)
		}
		slices.Sort(imports)

		text.WriteString("\nimport (\n")
		for _, it := range imports {
			text.WriteString(fmt.Sprintf("\t%#v", it))
		}
		text.WriteString("\n)\n")
	}

	var blocks [][2]string
	for name, block := range file.funcs {
		text := block.OutputText()
		blocks = append(blocks, [2]string{name, text})
	}
	slices.SortFunc(blocks, func(a, b [2]string) int { return cmp.Compare(a[0], b[0]) })

	for _, it := range blocks {
		text.WriteString("\n")
		text.WriteString(it[1])
		text.WriteString("\n")
	}

	return text.String()
}

func (blk *GoBlock) OutputText() string {
	text := strings.Builder{}

	if len(blk.header) > 0 {
		text.WriteString(blk.header)
		text.WriteString("\n")
	}

	if len(blk.vars) > 0 {
		text.WriteString("\tvar (\n")
		for _, it := range blk.vars {
			text.WriteString("\t")
			text.WriteString(it[0])
			text.WriteString(" ")
			text.WriteString(it[1])
			text.WriteString("\n")
		}
		text.WriteString("\t)\n")
	}

	for _, it := range blk.lines {
		text.WriteString(it)
		text.WriteString("\n")
	}

	if len(blk.footer) > 0 {
		text.WriteString(blk.footer)
		text.WriteString("\n")
	}

	return text.String()
}
