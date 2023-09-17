package hgl

import "github.com/qbart/hashira/hmath"

type Mesh struct {
	VertexData *VertexBuffer3f
	SubMeshes  []*SubMesh
}

type SubMesh struct {
	Model hmath.Matrix4
	UVs   *VertexBuffer2f
}
