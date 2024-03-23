package bot

import (
	"fmt"
	"os"
	"path"

	"axlab.dev/bit/input"
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
	if err := os.WriteFile(name, []byte(input.Text(text)), os.ModePerm); err != nil {
		Fatal(err, "failed to write text file `%s`", name)
	}
}
