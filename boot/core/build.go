package core

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

func NeedRebuild(output string, srcGlob string) (bool, error) {
	out, err := os.Stat(output)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return true, nil
		}
		return false, err
	}

	pre, pat := GlobParse(srcGlob)

	var srcTime time.Time
	if pat == "" {
		if stat, err := os.Stat(srcGlob); err == nil {
			srcTime = stat.ModTime()
		} else {
			return false, err
		}
	}

	if pre == "" {
		pre = "."
	}

	root, err := FS(pre)
	if err != nil {
		return false, err
	}

	srcFiles, errs := root.Glob(pat)
	if len(errs) == 0 {
		for _, src := range srcFiles {
			info, err := src.Info()
			if err != nil {
				errs = append(errs, err)
				continue
			}

			dt := info.ModTime()
			if srcTime.IsZero() || srcTime.Before(dt) {
				srcTime = dt
			}
		}
	}

	if len(errs) > 0 {
		return false, ErrorFromList(errs, "glob errors")
	}

	return out.ModTime().Before(srcTime), nil
}

var (
	projectRoot = projectRootFind()
)

func ProjectRoot() string {
	return projectRoot
}

func projectRootFind() string {
	cwd := Check(filepath.Abs("."))
	root := cwd
	for !projectRootCheck(root) {
		next := filepath.Join(root, "..")
		if next == "" || next == root {
			Fatal(fmt.Errorf("could not find project path [cwd=%s]", cwd))
		}
		root = next
	}
	return root
}

func projectRootCheck(path string) bool {
	make := filepath.Join(path, "maker.go")
	boot := filepath.Join(path, "boot")
	return IsFile(make) && IsDir(boot)
}
