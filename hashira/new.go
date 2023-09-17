package hashira

import (
	"github.com/qbart/hashira/ds"
	"github.com/qbart/hashira/hgl"
)

func New() *World {
	return &World{
		Maps: ds.NewHashMap[string, *Map](),
		Mesh: ds.NewHashMap[string, *hgl.Mesh](),
	}
}
