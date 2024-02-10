package common

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
)

const RegexpIgnoreCase = "(?i)"

func MatchesPattern(input, pattern string) bool {
	re := regexp.MustCompile(RegexpIgnoreCase + GlobRegex(pattern))
	return re.MatchString(input)
}

func GlobRegex(pattern string) string {
	var output []string

	next, runes := ' ', []rune(pattern)
	for len(runes) > 0 {
		next, runes = runes[0], runes[1:]
		switch next {
		case '/', '\\':
			output = append(output, `[/\\]`)
		case '?':
			output = append(output, `[^/\\]`)
		case '*':
			output = append(output, `[^/\\]*`)
		case '(', ')', '|':
			output = append(output, string(next))
		default:
			output = append(output, regexp.QuoteMeta(string(next)))
		}
	}
	return strings.Join(output, "")
}

func Glob(root, pattern string) (out []string) {
	root = Try(filepath.Abs(root))
	isPath := strings.Contains(pattern, "/")
	anchor := "^"
	if isPath {
		anchor = ""
	}

	re := regexp.MustCompile(RegexpIgnoreCase + anchor + "(" + GlobRegex(pattern) + ")$")
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		path = Relative(root, path)
		path = strings.Replace(path, "\\", "/", -1)

		var name string
		if isPath {
			name = path
		} else {
			name = d.Name()
		}

		if re.MatchString(name) {
			out = append(out, path)
		}
		return nil
	})
	return out
}

func Relative(base, path string) string {
	fullBase, err := filepath.Abs(base)
	NoError(err, "getting absolute base path for relative")

	fullPath, err := filepath.Abs(path)
	NoError(err, "getting absolute path for relative")

	rel, err := filepath.Rel(fullBase, fullPath)
	NoError(err, "getting relative path")
	return rel
}

func WithExtension(filename string, ext string) string {
	out := strings.TrimSuffix(filename, filepath.Ext(filename))
	return out + ext
}
