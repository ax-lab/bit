package bot

import (
	"fmt"
	"strings"

	"axlab.dev/bit/input"
)

func LexString(cur *input.Cursor) (length int, err error) {
	startPos := cur.Offset()
	delim := cur.ReadAny(`r"`, `r'`, `"`, `'`)
	if delim == "" {
		return 0, nil
	}

	raw := delim[0] == 'r'
	if raw {
		delim = delim[1:]
	}

	closed := false
	doubleDelim := delim + delim
	for !cur.Empty() {
		if !raw && cur.ReadString(`\`) {
			cur.Read()
		} else if !raw || !cur.ReadString(doubleDelim) {
			if cur.ReadString(delim) {
				closed = true
				break
			} else {
				cur.Read()
			}
		}
	}

	if !closed {
		err = fmt.Errorf("string literal missing closing `%s`", delim)
	}

	return cur.Offset() - startPos, err
}

func LexStringToString(str string) string {
	out := strings.Builder{}
	raw := false
	if strings.HasPrefix(str, "r") {
		str = str[1:]
		raw = true
	}

	delim, str, dbl := str[:1], str[1:], ""
	if raw {
		dbl = delim + delim
	}

	if strings.HasSuffix(str, delim) {
		str = str[:len(str)-1]
	}

	for len(str) > 0 {
		if raw {
			if pos := strings.Index(str, dbl); pos >= 0 {
				out.WriteString(str[:pos])
				out.WriteString(delim)
				str = str[pos+len(dbl):]
			} else {
				out.WriteString(str)
				str = ""
			}
		} else {
			if pos := strings.Index(str, "\\"); pos >= 0 {
				out.WriteString(str[:pos])
				str = str[pos+1:]
				if strings.HasPrefix(str, "\\") {
					out.WriteString("\\")
					str = str[1:]
				}
			} else {
				out.WriteString(str)
				str = ""
			}
		}
	}

	return out.String()
}
