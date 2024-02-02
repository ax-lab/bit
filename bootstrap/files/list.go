package files

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"axlab.dev/bit/logs"
)

type Info struct {
	Name  string
	Path  string
	Entry fs.DirEntry
}

func (info *Info) IsDir() bool {
	return info.Entry.IsDir()
}

func (info *Info) String() string {
	mode := "F"
	if info.IsDir() {
		mode = "D"
	}
	return fmt.Sprintf("%s %s", mode, info.Path)
}

func List(dirPath string) (out []Info) {
	dir := os.DirFS(dirPath)
	fs.WalkDir(dir, ".", func(entryPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			logs.Warn(err, "listing `%s`", dirPath)
			return nil
		}

		name := entry.Name()
		if name == "." {
			return nil
		}

		if isHidden := strings.HasPrefix(name, "."); isHidden {
			if entry.IsDir() {
				return fs.SkipDir
			} else {
				return nil
			}
		}

		out = append(out, Info{
			Name:  entry.Name(),
			Path:  path.Join(dirPath, entryPath),
			Entry: entry,
		})
		return nil
	})
	return out
}
