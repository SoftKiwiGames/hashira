package hashira

import (
	"github.com/qbart/hashira/ds"
	"github.com/qbart/hashira/hgl"
	"github.com/qbart/hashira/hmath"
)

type World struct {
	Resources *Resources
	Maps      *ds.HashMap[string, *Map]
	synced    bool
}

func (w *World) AddMap(name string, width int, height int, tileWidth int, tileHeight int) *Map {
	m := &Map{
		Width:              width,
		Height:             height,
		TileWidth:          tileWidth,
		TileHeight:         tileHeight,
		Layers:             ds.NewHashMap[string, *Layer](),
		SubMeshIndexByName: ds.NewHashMap[string, int](),
	}
	mesh := &hgl.Mesh{
		VertexData: hgl.NewVertexBuffer3f(m.VerticesNeeded()),
		SubMeshes:  make([]*hgl.SubMesh, 0),
	}
	m.Mesh = mesh
	w.Maps.Set(name, m)

	tw := float32(tileWidth)
	th := float32(tileHeight)

	for my := 0; my < height; my++ {
		for mx := 0; mx < width; mx++ {
			z := float32(0)
			i := (my*width + mx) * 6
			x := float32(i / 6 % width)
			y := float32(i / 6 / width)

			// first triangle
			//    2
			//  / |
			// 0--1
			//

			mesh.VertexData.Set(i+0, (x+0)*tw, (y+0)*th, z)
			mesh.VertexData.Set(i+1, (x+1)*tw, (y+0)*th, z)
			mesh.VertexData.Set(i+2, (x+1)*tw, (y+1)*th, z)

			// second triangle
			// 4--3
			// | /
			// 5
			mesh.VertexData.Set(i+3, (x+1)*tw, (y+1)*th, z)
			mesh.VertexData.Set(i+4, (x+0)*tw, (y+1)*th, z)
			mesh.VertexData.Set(i+5, (x+0)*tw, (y+0)*th, z)
		}
	}

	return m
}

func (w *World) AddLayer(mapName string, name string, z float32) *Layer {
	m := w.Maps.Get(mapName)
	subMesh := &hgl.SubMesh{
		Model: hmath.TranslationMatrix(hmath.Vertex{0, 0, z}),
		UVs:   hgl.NewVertexBuffer2f(m.VerticesNeeded()),
	}
	m.Mesh.SubMeshes = append(m.Mesh.SubMeshes, subMesh)
	m.SubMeshIndexByName.Set(name, len(m.Mesh.SubMeshes)-1)

	layer := &Layer{
		Z:    z,
		Data: make([][]int, m.Height),
	}
	for i := range layer.Data {
		layer.Data[i] = make([]int, m.Width)
	}
	m.Layers.Set(name, layer)

	return layer
}

func (w *World) AddLayerData(mapName string, name string, data [][]int) {
	m := w.Maps.Get(mapName)
	layer := m.Layers.Get(name)
	submesh := m.Mesh.SubMeshes[m.SubMeshIndexByName.Get(name)]

	for my := 0; my < m.Height; my++ {
		for mx := 0; mx < m.Width; mx++ {
			layer.Data[my][mx] = data[my][mx]
			if w.synced {
				w.buildTileUV(m, submesh, mx, my, layer.Data[my][mx])
			}
		}
	}
}

func (w *World) SetTile(mapName string, layerName string, x, y int, tile int) {
	m := w.Maps.Get(mapName)
	layer := m.Layers.Get(layerName)
	layer.SetTile(x, y, tile)

	layerMesh := m.Mesh.SubMeshes[m.SubMeshIndexByName.Get(layerName)]
	w.buildTileUV(m, layerMesh, x, y, tile)
}

func (w *World) buildTileUV(m *Map, s *hgl.SubMesh, x, y int, tile int) {
	// flip y for natural top down order
	y = m.Height - y - 1
	i := (y*int(m.Width) + x) * 6

	u, v, u2, v2 := w.Resources.GetTileset().TextureUV(tile, m.TileWidth, m.TileHeight)

	// first triangle
	//    2
	//  / |
	// 0--1
	//
	s.UVs.Set(i+0, u, v2)
	s.UVs.Set(i+1, u2, v2)
	s.UVs.Set(i+2, u2, v)
	// second triangle
	// 4--3
	// | /
	// 5
	s.UVs.Set(i+3, u2, v)
	s.UVs.Set(i+4, u, v)
	s.UVs.Set(i+5, u, v2)
}

func (w *World) Resync() {
	w.synced = false
}

// Sync rebuilds all meshes and UVs when needed
// since maps and layers can be added before tileset is loaded.
func (w *World) Sync() {
	if w.synced {
		return
	}
	w.Maps.ForEach(func(_ string, m *Map) {
		m.Layers.ForEach(func(layerName string, layer *Layer) {
			index := m.SubMeshIndexByName.Get(layerName)
			layerMesh := m.Mesh.SubMeshes[index]

			for my := 0; my < m.Height; my++ {
				for mx := 0; mx < m.Width; mx++ {
					tile := layer.Tile(mx, my)
					w.buildTileUV(m, layerMesh, mx, my, tile)
				}
			}
		})
	})

	w.synced = true
}
