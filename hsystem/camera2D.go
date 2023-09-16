package hsystem

import "github.com/qbart/hashira/hmath"

type Camera2D struct {
	ViewMatrix hmath.Matrix4
	Position   hmath.Vertex
}

func (c *Camera2D) Translate(x, y float32) {
	c.Position[0] = -x
	c.Position[1] = -y
	c.Position[2] = 0
	c.ViewMatrix = hmath.TranslationMatrix(c.Position)
}
