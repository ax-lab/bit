package core

import (
	"fmt"
	"strings"
)

const (
	IdEscapePrefix = "x__"
	IdEscapeMiddle = "__u"
)

func GetPlainIdentifier(id string) string {
	if id == "" {
		return IdEscapePrefix
	}

	ok := true
	for n, chr := range id {
		if n == 0 {
			ok = chr == '_' || IsAlpha(chr)
		} else {
			ok = chr == '_' || IsAlphaNum(chr)
		}
		if !ok {
			break
		}
	}

	if ok {
		if idBlacklistMap[id] || strings.HasPrefix(id, IdEscapePrefix) {
			return IdEscapePrefix + id
		} else {
			return id
		}
	}

	out := strings.Builder{}
	out.WriteString(IdEscapePrefix)
	txt := id

outer:
	for len(txt) > 0 {
		if esc := IdEscapeMiddle; strings.HasPrefix(txt, esc) {
			out.WriteString(esc)
			out.WriteString(esc)
			txt = txt[len(esc):]
		} else {
			for n, chr := range txt {
				if n == 0 {
					if IsAlphaNum(chr) {
						out.WriteRune(chr)
					} else {
						out.WriteString(fmt.Sprintf("%s%04X", esc, int(chr)))
					}
				} else {
					txt = txt[n:]
					continue outer
				}
			}
			txt = ""
		}
	}

	return out.String()
}

var idBlacklist = []string{
	IdEscapePrefix,

	// Go identifiers
	"break",
	"case",
	"chan",
	"const",
	"continue",
	"default",
	"defer",
	"else",
	"fallthrough",
	"for",
	"func",
	"go",
	"goto",
	"if",
	"import",
	"interface",
	"map",
	"package",
	"range",
	"return",
	"select",
	"struct",
	"switch",
	"type",
	"var",

	// C++ identifiers
	"alignas",
	"alignof",
	"and",
	"and_eq",
	"asm",
	"atomic_cancel",
	"atomic_commit",
	"atomic_noexcept",
	"auto",
	"bitand",
	"bitor",
	"bool",
	"break",
	"case",
	"catch",
	"char",
	"char8_t",
	"char16_t",
	"char32_t",
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
	"decltype",
	"default",
	"delete",
	"do",
	"double",
	"dynamic_cast",
	"else",
	"enum",
	"explicit",
	"export",
	"extern",
	"false",
	"float",
	"for",
	"friend",
	"goto",
	"if",
	"inline",
	"int",
	"long",
	"mutable",
	"namespace",
	"new",
	"noexcept",
	"not",
	"not_eq",
	"nullptr",
	"operator",
	"or",
	"or_eq",
	"private",
	"protected",
	"public",
	"reflexpr",
	"register",
	"reinterpret_cast",
	"requires",
	"return",
	"short",
	"signed",
	"sizeof",
	"static",
	"static_assert",
	"static_cast",
	"struct",
	"switch",
	"synchronized",
	"template",
	"this",
	"thread_local",
	"throw",
	"true",
	"try",
	"typedef",
	"typeid",
	"typename",
	"union",
	"unsigned",
	"using",
	"virtual",
	"void",
	"volatile",
	"wchar_t",
	"while",
	"xor",
	"xor_eq",

	// Javascript keywords (https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar)
	"abstract",
	"arguments",
	"as",
	"async",
	"await",
	"boolean",
	"break",
	"byte",
	"case",
	"catch",
	"char",
	"class",
	"const",
	"continue",
	"debugger",
	"default",
	"delete",
	"do",
	"double",
	"else",
	"enum",
	"eval",
	"export",
	"extends",
	"false",
	"final",
	"finally",
	"float",
	"for",
	"from",
	"function",
	"get",
	"goto",
	"if",
	"implements",
	"import",
	"in",
	"instanceof",
	"int",
	"interface",
	"let",
	"long",
	"native",
	"new",
	"null",
	"of",
	"package",
	"private",
	"protected",
	"public",
	"return",
	"set",
	"short",
	"static",
	"super",
	"switch",
	"synchronized",
	"this",
	"throw",
	"throws",
	"transient",
	"true",
	"try",
	"typeof",
	"var",
	"void",
	"volatile",
	"while",
	"with",
	"yield",
}

var idBlacklistMap = (func() (out map[string]bool) {
	out = make(map[string]bool)
	for _, it := range idBlacklist {
		out[it] = true
	}
	return out
}())
