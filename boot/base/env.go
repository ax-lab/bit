package base

import (
	"os"
	"path/filepath"
	"runtime"
)

var projectDir = (func() string {
	dir := filepath.Dir(CurrentFile())
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

// Returns the absolute file path for the source file of the caller function.
func CurrentFile() string {
	_, callerFile, _, hasInfo := runtime.Caller(1)
	if !hasInfo {
		panic("could not retrieve caller file name")
	}
	if !filepath.IsAbs(callerFile) {
		panic("caller file name is not an absolute path")
	}
	return filepath.Clean(callerFile)
}

// Returns true if the given path is a directory.
func IsDir(path string) bool {
	if stat, err := os.Stat(path); err == nil {
		return stat.IsDir()
	}
	return false
}

// Returns true if the given path is a file.
func IsFile(path string) bool {
	if stat, err := os.Stat(path); err == nil {
		return !stat.IsDir()
	}
	return false
}
