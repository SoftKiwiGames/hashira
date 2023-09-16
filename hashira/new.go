package hashira

import "github.com/qbart/hashira/ds"

func New() *World {
	return &World{
		Maps: ds.NewHashMap[string, *Map](),
	}
}
