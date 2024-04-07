package core

import (
	"errors"
	"io/fs"
	"os"
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
		return false, Errors(errs, "glob errors")
	}

	return out.ModTime().Before(srcTime), nil
}
