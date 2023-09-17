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
