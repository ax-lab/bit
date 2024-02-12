package files

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"axlab.dev/bit/common"
)

func WorkingDir() string {
	return common.Try(filepath.Abs("."))
}

func IsDir(path string) bool {
	if stat, err := os.Stat(path); err == nil {
		return stat.IsDir()
	}
	return false
}

func MakeDir(path string) {
	common.Check(os.MkdirAll(path, fs.ModePerm))
}

func IsFile(path string) bool {
	if stat, err := os.Stat(path); err == nil {
		return !stat.IsDir()
	}
	return false
}

var projectDir = (func() string {
	dir := filepath.Dir(FileName())
	dir, valid := filepath.Join(dir, "..", ".."), false
	for {
		if IsDir(filepath.Join(dir, "bootstrap/bit")) && IsFile(filepath.Join(dir, "go.work")) {
			valid = true
			break
		} else {
			next := filepath.Dir(dir)
			if next == dir || next == "." {
				break
			}
		}
	}
	if !valid {
		panic("could not find bootstrap project dir")
	}
	return dir
})()

// Returns the absolute root directory for the project.
func ProjectDir() string {
	return projectDir
}

// Returns the Go filename of the caller function.
func FileName() string {
	_, callerFile, _, hasInfo := runtime.Caller(1)
	if !hasInfo {
		log.Fatal("could not retrieve caller file name")
	}
	if !filepath.IsAbs(callerFile) {
		log.Fatal("caller file name is not an absolute path")
	}
	return filepath.Clean(callerFile)
}
