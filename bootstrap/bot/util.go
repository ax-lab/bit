package bot

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"unicode"
)

func Fatal(err error, msg string, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	fmt.Fprintf(os.Stderr, "\nFATAL: %s -- %v\n\n", msg, err)
	os.Exit(1)
}

func ReadText(name string) string {
	data, err := os.ReadFile(name)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		Fatal(err, "failed to read text file `%s`", name)
	}
	return string(data)
}

func WriteText(name, text string) {
	if dir := path.Dir(name); dir != "." {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			Fatal(err, "failed to create directory for text file `%s`", name)
		}
	}
	if err := os.WriteFile(name, []byte(Text(text)), os.ModePerm); err != nil {
		Fatal(err, "failed to write text file `%s`", name)
	}
}

func Text(input string) string {
	lines := Lines(TrimEnd(input))
	if len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}

	if len(lines) > 0 {
		line := lines[0]
		trim := TrimSta(line)
		if diff := line[:len(line)-len(trim)]; diff != "" {
			for i, it := range lines {
				if strings.HasPrefix(it, diff) {
					lines[i] = it[len(diff):]
				}
			}
		}
	}

	if len(lines) == 0 || lines[len(lines)-1] != "" {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func Lines(input string) []string {
	return reLines.Split(input, -1)
}

func TrimSta(input string) string {
	return strings.TrimLeftFunc(input, IsSpace)
}

func TrimEnd(input string) string {
	return strings.TrimRightFunc(input, IsSpace)
}

func IsSpace(chr rune) bool {
	return chr != '\r' && chr != '\n' && unicode.IsSpace(chr)
}

var (
	reLines = regexp.MustCompile(`\r?\n`)
)
