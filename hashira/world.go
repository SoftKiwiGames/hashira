package hashira

import (
	"github.com/qbart/hashira/ds"
	"github.com/qbart/hashira/hgl"
	"github.com/qbart/hashira/hmath"
)

type World struct {
	Maps *ds.HashMap[string, *Map]
	Mesh *ds.HashMap[string, *hgl.Mesh]
}

func (w *World) AddMap(name string, width int, height int) *Map {
	m := &Map{
		Width:   width,
		Height:  height,
		Layers:  ds.NewHashMap[string, *Layer](),
		SubMesh: ds.NewHashMap[string, *hgl.SubMesh](),
	}
	w.Maps.Set(name, m)

	mesh := &hgl.Mesh{
		VertexData: hgl.NewVertexBuffer3f(m.VerticesNeeded()),
		SubMeshes:  make([]*hgl.SubMesh, 0),
	}
	w.Mesh.Set(name, mesh)

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

			mesh.VertexData.Set(i+0, x, y, z)
			mesh.VertexData.Set(i+1, x+1, y, z)
			mesh.VertexData.Set(i+2, x+1, y+1, z)

			// second triangle
			// 4--3
			// | /
			// 5
			mesh.VertexData.Set(i+3, x+1, y+1, z)
			mesh.VertexData.Set(i+4, x, y+1, z)
			mesh.VertexData.Set(i+5, x, y, z)
		}
	}

	return m
}

func (w *World) AddLayer(mapName string, name string, z float32) *Layer {
	m := w.Maps.Get(mapName)
	mesh := w.Mesh.Get(mapName)
	subMesh := &hgl.SubMesh{
		Model: hmath.TranslationMatrix(hmath.Vertex{0, 0, z}),
		UVs:   hgl.NewVertexBuffer2f(m.VerticesNeeded()),
	}
	mesh.SubMeshes = append(mesh.SubMeshes, subMesh)

	layer := &Layer{
		Z: z,
	}
	m.Layers.Set(name, layer)
	m.SubMesh.Set(name, subMesh)

	return layer
}

func (w *World) AddLayerData(mapName string, name string, data [][]int) {
	m := w.Maps.Get(mapName)
	layer := m.Layers.Get(name)
	layerMesh := m.SubMesh.Get(name)
	layerData := make([][]int, m.Height)

	for my := 0; my < m.Height; my++ {
		data[my] = make([]int, m.Width)
		for mx := 0; mx < m.Width; mx++ {
			tile := layer.Tile(mx, m.Height-my-1)
			tmp := Tileset{
				TileSize:      16,
				TextureWidth:  256,
				TextureHeight: 256,
			}
			SetTileAt(m, layerMesh, &tmp, mx, my, tile)
		}
	}
	layer.Data = layerData
}

func TileUV(tile int, tileSize int, tilesetWidth int, tilesetHeight int) (float32, float32, float32, float32) {
	tilesPerRow := tilesetWidth / tileSize
	rowX := tile % tilesPerRow
	rowY := tile / tilesPerRow

	u := float32(rowX*tileSize) / float32(tilesetHeight)
	u2 := float32((rowX+1)*tileSize) / float32(tilesetWidth)
	v := float32(rowY*tileSize) / float32(tilesetHeight)
	v2 := float32((rowY+1)*tileSize) / float32(tilesetHeight)

	return u, v, u2, v2
}

func SetTileAt(m *Map, s *hgl.SubMesh, tileset *Tileset, x, y int, tile int) {
	i := (y*int(m.Width) + x) * 6

	u, v, u2, v2 := TileUV(tile, tileset.TileSize, tileset.TextureWidth, tileset.TextureHeight)

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
