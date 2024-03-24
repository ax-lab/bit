package bot

import (
	"strings"

	"axlab.dev/bit/input"
)

const (
	COMMENT_CHAR      = '#'
	COMMENT_DELIM     = '/'
	COMMENT_DELIM_STA = "/#"
	COMMENT_DELIM_END = "#/"
)

func LexComment(cursor *input.Cursor) int {
	noComment := func(chr rune) bool {
		return chr != COMMENT_CHAR && chr != COMMENT_DELIM
	}

	startPos := cursor.Offset()
	if next := cursor.Peek(); next == COMMENT_CHAR {
		cursor.SkipWhile(func(chr rune) bool { return chr != '\n' && chr != '\r' })
		return cursor.Offset() - startPos
	} else if next != COMMENT_DELIM {
		return 0
	}

	count := 0
	for !cursor.Empty() {
		next := cursor.Text()
		if strings.HasPrefix(next, COMMENT_DELIM_STA) {
			cursor.Advance(len(COMMENT_DELIM_STA))
			count += 1
		} else if strings.HasPrefix(next, COMMENT_DELIM_END) {
			cursor.Advance(len(COMMENT_DELIM_END))
			count -= 1
			if count == 0 {
				break
			}
		} else {
			cursor.Advance(1)
			cursor.SkipWhile(noComment)
		}
	}

	return cursor.Offset() - startPos
}
