package files

import (
	"path"
	"strings"
)

func PathNew[T ~string](args ...T) Path {
	out := pathBuilder{}
	for _, it := range args {
		out.Reserve(string(it))
	}

	if !out.Init() {
		return ""
	}

	for _, it := range args[out.Skip():] {
		out.Push(string(it))
	}

	return out.Done()
}

type Path string

func (p Path) Len() int {
	return len(p)
}

func (p Path) Clean() Path {
	return PathNew(p)
}

func (p Path) HasPrefix(prefix Path) bool {
	preLen := prefix.Len()
	if preLen == 0 {
		return true
	}
	return strings.HasPrefix(string(p), string(prefix)) && (p.Len() == preLen || p[preLen] == '/')
}

func (p Path) IsRoot() bool {
	return len(p) > 0 && p[0] == '/'
}

func (p Path) IsOutside() bool {
	return strings.HasPrefix(string(p), "..") && (len(p) == 2 || p[2] == '/')
}

func (p Path) Push(elems ...string) Path {
	out := pathBuilder{}

	out.Reserve(string(p))
	for _, it := range elems {
		out.Reserve(it)
	}

	if !out.Init() {
		return ""
	}

	out.Push(string(p))
	for _, it := range elems {
		out.Push(it)
	}

	return out.Done()
}

func (p Path) Join(elems ...Path) Path {
	out := pathBuilder{}

	out.Reserve(string(p))
	for _, it := range elems {
		out.Reserve(string(it))
	}

	if !out.Init() {
		return ""
	}

	out.Push(string(p))
	for _, it := range elems {
		out.Push(string(it))
	}

	return out.Done()
}

// Last element of the path, if any.
func (p Path) Base() string {
	out := path.Base(string(p))
	if out == "/" || out == "." {
		return ""
	}
	return out
}

// File name extension including the dot, if any.
func (p Path) Ext() string {
	return path.Ext(string(p))
}

// Path with the last element removed. Returns `\` or `.` if empty.
func (p Path) Dir() Path {
	dir := path.Dir(string(p))
	return Path(dir)
}

type pathBuilder struct {
	skip int
	size int
	elem int
	path []byte
}

func (pb *pathBuilder) Reserve(elem string) {
	if len(elem) == 0 {
		return
	}
	if rooted := elem[0] == '/' || elem[0] == '\\'; rooted {
		pb.skip += pb.elem
		pb.size, pb.elem = 0, 0
	}
	pb.size += len(elem)
	pb.elem += 1
}

func (pb *pathBuilder) Skip() int {
	return pb.skip
}

func (pb *pathBuilder) Init() bool {
	if pb.size == 0 {
		return false
	}
	pb.size += pb.elem - 1
	pb.path = make([]byte, 0, pb.size)
	return true
}

func (pb *pathBuilder) Push(elem string) {
	if len(elem) == 0 {
		return
	}

	if len(pb.path) > 0 {
		pb.path = append(pb.path, '/')
	}

	rest := string(elem)
	for len(rest) > 0 {
		if sep := strings.IndexRune(rest, '\\'); sep >= 0 {
			pb.path = append(pb.path, rest[:sep]...)
			pb.path = append(pb.path, '/')
			rest = rest[sep+1:]
		} else {
			pb.path = append(pb.path, rest...)
			rest = ""
		}
	}
}

func (pb *pathBuilder) Done() Path {
	if pb.size < len(pb.path) {
		panic("PathNew: invalid size calculation")
	}

	out := path.Clean(string(pb.path))
	return Path(out)
}
