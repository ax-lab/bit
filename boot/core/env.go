package core

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	projectRoot = projectRootFind()
	bootDir     = "boot"
	rootFiles   = []string{
		"bit",
		"Makefile",
		"go.work",
	}
)

func ProjectRoot() string {
	return projectRoot
}

func WorkingDir() string {
	return Try(filepath.Abs("."))
}

func IsDir(path string) bool {
	if stat, err := os.Stat(path); err == nil {
		return stat.IsDir()
	}
	return false
}

func IsFile(path string) bool {
	if stat, err := os.Stat(path); err == nil {
		return !stat.IsDir()
	}
	return false
}

func CleanPath(name string) (out string, err error) {
	out = path.Clean(name)

	valid := true
	if out == "" || strings.ContainsAny(out, "\"\r\n\t\\*?:|<>") || strings.ContainsFunc(out, func(r rune) bool { return r < 32 }) {
		valid = false
	} else {
		for _, it := range strings.Split(out, "/") {
			if it == "." || it == ".." {
				valid = false
				break
			}
		}
	}

	if !valid {
		return "", fmt.Errorf("invalid path: %#v", name)
	} else if strings.HasPrefix(out, "/") {
		return "", fmt.Errorf("absolute path is not allowed: %s", name)
	}

	return out, nil
}

func projectRootFind() string {
	cwd := WorkingDir()
	root := cwd
	for !projectRootCheck(root) {
		next := filepath.Join(root, "..")
		if next == "" || next == root {
			panic(fmt.Sprintf("could not find project path [cwd=%s]", cwd))
		}
		root = next
	}
	return root
}

func projectRootCheck(path string) bool {
	for _, it := range rootFiles {
		full := filepath.Join(path, it)
		if !IsFile(full) {
			return false
		}
	}

	boot := filepath.Join(path, bootDir)
	return IsDir(boot)
}
