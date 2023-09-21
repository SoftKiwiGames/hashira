package hashira

import "github.com/qbart/hashira/hmath"

type Camera2D struct {
	ViewMatrix hmath.Matrix4
	Position   hmath.Vertex
	Zoom       float32
}

func (c *Camera2D) Translate(x, y float32) {
	c.Position[0] = -x
	c.Position[1] = -y
	c.Position[2] = 0
	c.ViewMatrix = hmath.TranslationMatrix(c.Position)
}

func (c *Camera2D) Projection(canvasWidth int, canvasHeight int, tileSize int) hmath.Matrix4 {
	return hmath.Ortho(
		-float32((canvasWidth)/tileSize)/(2*c.Zoom),
		float32((canvasWidth)/tileSize)/(2*c.Zoom),
		-float32((canvasHeight)/tileSize)/(2*c.Zoom),
		float32((canvasHeight)/tileSize)/(2*c.Zoom),
		-100, 100,
	)
}
