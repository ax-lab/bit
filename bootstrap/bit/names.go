package bit

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
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

func EncodeIdentifier(name string) string {
	if name == "" {
		return "_$"
	}

	out := strings.Builder{}
	valid := true
	for n, chr := range name {
		if n == 0 && IsDigit(chr) {
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
	return chr == '_' || IsDigit(chr) || IsAlpha(chr)
}

func EscapeNameChar(chr rune, out *strings.Builder) {
	switch chr {
	case '-':
		out.WriteString("_$")
	default:
		out.WriteString(fmt.Sprintf("_$u%04X_", chr))
	}
}
