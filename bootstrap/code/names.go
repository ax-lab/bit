package code

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"

	"axlab.dev/bit/common"
)

type NameMap struct {
	mutex sync.Mutex
	root  NameScope
}

func (m *NameMap) DeclareGlobal(name string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.root.declare(name)
}

func (m *NameMap) NewChild() *NameScope {
	return &NameScope{
		nameMap: m,
		parent:  &m.root,
		names:   nil,
	}
}

type NameScope struct {
	nameMap *NameMap
	parent  *NameScope
	names   map[string]bool
}

func (m *NameScope) NewChild() *NameScope {
	return &NameScope{
		nameMap: m.nameMap,
		parent:  m,
		names:   nil,
	}
}

func (m *NameScope) DeclareUnique(name string) string {
	m.nameMap.mutex.Lock()
	defer m.nameMap.mutex.Unlock()
	return m.declareUnique(name)
}

func (m *NameScope) declare(name string) bool {
	if _, ok := m.names[name]; ok {
		return false
	}

	if m.names == nil {
		m.names = make(map[string]bool)
	}
	m.names[name] = true
	return true
}

func (m *NameScope) exists(name string) bool {
	if _, ok := m.names[name]; ok {
		return true
	}
	if m.parent != nil {
		return m.parent.exists(name)
	}
	return false
}

func (m *NameScope) declareUnique(name string) string {
	name = m.getUnique(name)
	m.declare(name)
	return name
}

func (m *NameScope) getUnique(name string) string {
	for counter := 0; counter < 1000; counter++ {
		cur := name
		if counter > 100 {
			cur = fmt.Sprintf("%s_$$%06X", name, rand.Int())
		} else if counter > 0 {
			cur = fmt.Sprintf("%s_$$%d", name, counter)
		}
		if !m.exists(cur) {
			return cur
		}
	}
	panic(fmt.Sprintf("failed to generate unique name: %s", name))
}

var reservedName = (func() map[string]bool {
	out := make(map[string]bool)
	for _, it := range reservedNameList() {
		out[it] = true
	}
	return out
})()

func EncodeIdentifier(name string) string {
	if reservedName[name] {
		return "_$" + name
	}

	out := strings.Builder{}
	valid := true
	for n, chr := range name {
		if n == 0 && common.IsDigit(chr) {
			valid = false
			out.WriteString("_$")
			out.WriteRune(chr)
			continue
		}

		chrValid := ValidNameChar(chr)
		if valid {
			if !chrValid {
				valid = false
				out.WriteString(name[:n])
				EscapeNameChar(chr, &out)
			}
		} else if chrValid {
			out.WriteRune(chr)
		} else {
			EscapeNameChar(chr, &out)
		}
	}

	if valid {
		return name
	}

	return out.String()
}

func ValidNameChar(chr rune) bool {
	return chr == '_' || common.IsDigit(chr) || common.IsAlpha(chr)
}

func EscapeNameChar(chr rune, out *strings.Builder) {
	switch chr {
	case '-':
		out.WriteString("_$")
	default:
		out.WriteString(fmt.Sprintf("_$u%04X_", chr))
	}
}

func reservedNameList() []string {
	return []string{
		"",
		"_",
		"abstract",
		"alignas",
		"alignof",
		"and",
		"and_eq",
		"as",
		"asm",
		"assert",
		"async",
		"atomic_cancel",
		"atomic_commit",
		"atomic_noexcept",
		"auto",
		"await",
		"become",
		"bitand",
		"bitor",
		"bool",
		"box",
		"break",
		"case",
		"catch",
		"chan",
		"char",
		"char16_t",
		"char32_t",
		"char8_t",
		"class",
		"compl",
		"concept",
		"const",
		"consteval",
		"constexpr",
		"constinit",
		"const_cast",
		"continue",
		"co_await",
		"co_return",
		"co_yield",
		"crate",
		"debugger",
		"decltype",
		"def",
		"default",
		"defer",
		"del",
		"delete",
		"do",
		"double",
		"dyn",
		"dynamic_cast",
		"elif",
		"else",
		"enum",
		"except",
		"explicit",
		"export",
		"extends",
		"extern",
		"fallthrough",
		"false",
		"False",
		"final",
		"finally",
		"float",
		"fn",
		"for",
		"friend",
		"from",
		"func",
		"function",
		"global",
		"go",
		"goto",
		"if",
		"impl",
		"import",
		"in",
		"inline",
		"instanceof",
		"int",
		"interface",
		"is",
		"lambda",
		"let",
		"long",
		"loop",
		"macro",
		"map",
		"match",
		"mod",
		"move",
		"mut",
		"mutable",
		"namespace",
		"new",
		"noexcept",
		"None",
		"nonlocal",
		"not",
		"not_eq",
		"null",
		"nullptr",
		"operator",
		"or",
		"or_eq",
		"override",
		"package",
		"pass",
		"priv",
		"private",
		"protected",
		"pub",
		"public",
		"raise",
		"range",
		"ref",
		"reflexpr",
		"register",
		"reinterpret_cast",
		"requires",
		"return",
		"select",
		"self",
		"Self",
		"short",
		"signed",
		"sizeof",
		"static",
		"static_assert",
		"static_cast",
		"struct",
		"super",
		"switch",
		"synchronized",
		"template",
		"this",
		"thread_local",
		"throw",
		"trait",
		"true",
		"True",
		"try",
		"type",
		"typedef",
		"typeid",
		"typename",
		"typeof",
		"union",
		"unsafe",
		"unsigned",
		"unsized",
		"use",
		"using",
		"var",
		"virtual",
		"void",
		"volatile",
		"wchar_t",
		"where",
		"while",
		"with",
		"xor",
		"xor_eq",
		"yield",
	}
}
