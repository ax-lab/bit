package core

import (
	"cmp"
	"fmt"
	"strings"
)

type LocationPos struct {
	File   string
	StaCol int
	StaRow int
	EndCol int
	EndRow int
}

func Location(file string, pos ...int) LocationPos {
	var staRow, staCol, endRow, endCol int

	valid := true
	for i, it := range pos {
		if it < 0 {
			valid = false
		}
		switch i {
		case 0:
			staRow = it
		case 1:
			staCol = it
		case 2:
			endRow = it
		case 3:
			endCol = it
		default:
			valid = false
		}

		if !valid {
			break
		}
	}

	valid = valid &&
		((endRow == 0 || endRow >= staRow) &&
			(endCol == 0 || endCol >= staCol || endRow != 0))

	if !valid {
		panic("Location: invalid position")
	}

	return LocationPos{
		File:   file,
		StaCol: staCol,
		StaRow: staRow,
		EndCol: endCol,
		EndRow: endRow,
	}
}

func (loc LocationPos) String() string {
	out := strings.Builder{}
	if loc.File != "" {
		out.WriteString(loc.File)
	}

	if loc.StaRow > 0 {
		if out.Len() == 0 {
			out.WriteString("@ ")
		} else {
			out.WriteString(" @ ")
		}

		out.WriteString(fmt.Sprintf("L%03d", loc.StaRow))
		if loc.StaCol > 0 {
			out.WriteString(fmt.Sprintf(":%02d", loc.StaCol))
		}

		if loc.EndRow > loc.StaRow {
			out.WriteString(fmt.Sprintf("…L%03d", loc.EndRow))
			if loc.EndCol > 0 {
				out.WriteString(fmt.Sprintf(":%02d", loc.EndCol))
			}
		} else if (loc.EndRow == 0 || loc.EndRow == loc.StaRow) && loc.EndCol > loc.StaCol {
			out.WriteString(fmt.Sprintf("…%02d", loc.EndCol))
		}
	}

	return out.String()
}

func (loc LocationPos) Valid() bool {
	return loc.File != "" || loc.StaRow > 0
}

func (loc LocationPos) Compare(other LocationPos) int {
	if res := cmp.Compare(loc.File, other.File); res != 0 {
		return res
	}

	if res := cmp.Compare(loc.StaRow, other.StaRow); res != 0 {
		return res
	}

	if res := cmp.Compare(loc.StaCol, other.StaCol); res != 0 {
		return res
	}

	if res := cmp.Compare(loc.EndRow, other.EndRow); res != 0 {
		return res
	}

	if res := cmp.Compare(loc.EndCol, other.EndCol); res != 0 {
		return res
	}

	return 0
}
