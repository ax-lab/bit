package cpp

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

func StringLiteral(str string) string {
	out := strings.Builder{}
	WriteStringLiteral(str, &out)
	return out.String()
}

func WriteStringLiteral(str string, out io.StringWriter) error {
	if _, err := out.WriteString("\""); err != nil {
		return err
	}
	for _, chr := range str {
		if err := cppOutputChar(chr, out); err != nil {
			return err
		}
	}
	if _, err := out.WriteString("\""); err != nil {
		return err
	}
	return nil
}

func cppOutputChar(chr rune, out io.StringWriter) error {
	seq := ""
	switch chr {
	case '?':
		seq = "\\?"
	case '"':
		seq = "\\\""
	case '\'':
		seq = "\\'"
	case '\\':
		seq = "\\\\"
	case '\x00':
		seq = "\\0"
	case '\t':
		seq = "\\t"
	case '\n':
		seq = "\\n"
	case '\r':
		seq = "\\r"
	case '\x08':
		seq = "\\b"
	default:
		if cppIsSafeStrChar(chr) {
			if _, err := out.WriteString(string(chr)); err != nil {
				return err
			}
		} else {
			buf := [utf8.UTFMax]byte{}
			len := utf8.EncodeRune(buf[:], chr)
			for _, b := range buf[:len] {
				if _, err := out.WriteString(fmt.Sprintf("\\x%X", b)); err != nil {
					return err
				}
			}
		}
	}
	if seq != "" {
		if _, err := out.WriteString(seq); err != nil {
			return err
		}
	}

	return nil
}

func cppIsSafeStrChar(chr rune) bool {
	switch chr {
	case
		'_', ' ', '!', '#', '$', '%', '&', '(', ')', '*', '+', ',', '-', '.', '/',
		':', ';', '<', '=', '>', '@', '[', ']', '^', '`', '{', '|', '}', '~':
		return true
	}

	if 'A' <= chr && chr <= 'Z' {
		return true
	}

	if 'a' <= chr && chr <= 'z' {
		return true
	}

	if '0' <= chr && chr <= '9' {
		return true
	}

	return false
}
