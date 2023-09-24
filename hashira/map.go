package hashira

import (
	"github.com/qbart/hashira/ds"
	"github.com/qbart/hashira/hgl"
)

type Map struct {
	Width      int
	Height     int
	TileWidth  int
	TileHeight int

	Layers  *ds.HashMap[string, *Layer]
	SubMesh *ds.HashMap[string, *hgl.SubMesh]
}

func (m *Map) Center() (x, y float32) {
	return float32(m.Width) / 2, float32(m.Height) / 2
}

func (m *Map) VerticesNeeded() int {
	// map size * 6 vertices per tile (we could share vertices between tiles but this is easier)
	return m.Width * m.Height * 6
}
