package code

import "sync"

type TypeKey interface {
	Types() []Type
}

type typeKeyLookup struct {
	sync sync.Mutex
	head typeKey
	rest map[*typeData]*typeKeyLookup
}

func (keyLookup *typeKeyLookup) Get(list []Type, offset int) *typeKey {
	if offset < 0 || offset > len(list) {
		panic("invalid offset")
	}

	if len(list) == offset {
		return &keyLookup.head
	}

	keyLookup.sync.Lock()
	defer keyLookup.sync.Unlock()
	if keyLookup.rest == nil {
		keyLookup.rest = make(map[*typeData]*typeKeyLookup)
	}

	data := list[offset].data
	rest := keyLookup.rest[data]
	if rest == nil {
		rest = &typeKeyLookup{
			head: typeKey{list[:offset+1]},
		}
		keyLookup.rest[data] = rest
	}

	return rest.Get(list, offset+1)
}

type typeKey struct {
	list []Type
}

func (key typeKey) Types() []Type {
	return key.list
}
