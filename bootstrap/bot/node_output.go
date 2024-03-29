package bot

import "fmt"

func (ls NodeList) GoType() GoType {
	last := ls.Len() - 1
	if last >= 0 {
		node, ok := ls.Get(last).(GoCode)
		if ok {
			return node.GoType()
		}
	}
	return GoTypeNone
}

func (ls NodeList) GoOutputOne(blk *GoBlock) (out GoVar, err error) {
	if cnt := ls.Len(); cnt != 1 {
		err = fmt.Errorf("invalid list arity: %d", cnt)
		return
	}

	code, ok := ls.Get(0).(GoCode)
	if !ok {
		err = fmt.Errorf("cannot output `%s` as Go code", ls.Get(0).Repr())
		return
	}

	return code.GoOutput(blk), nil
}

func (ls NodeList) GoOutputAll(blk *GoBlock) (out GoVar) {
	nodes := ls.Slice()
	for _, it := range nodes {
		code, ok := it.(GoCode)
		if !ok {
			blk.AddError(it.Span().NewError("cannot output `%s` as Go code", it.Repr()))
		} else {
			out = code.GoOutput(blk)
		}

		if blk.HasErrors() {
			break
		}
	}

	return out
}
