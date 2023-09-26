package hashira

import (
	"github.com/qbart/hashira/hmath"
)

type Camera2D struct {
	ViewMatrix hmath.Matrix4
	Position   hmath.Vertex
	Zoom       float32
}

func (c *Camera2D) ZoomBy(delta float32) {
	c.Zoom = hmath.Clamp(c.Zoom+delta, 0.5, 20)
	// c.Zoom = hmath.Clamp(c.Zoom+delta, 1, 20)

	// correction for going from 0.5 to 1
	if hmath.CloseTo(c.Zoom, 1.5, 0.1) {
		c.Zoom = 1
	}
}

func (c *Camera2D) SetZoom(zoom float32) {
	c.Zoom = zoom
}

func (c *Camera2D) Translate(x, y float32) {
	c.Position[0] = -x
	c.Position[1] = -y
	c.Position[2] = 0
	c.ViewMatrix = hmath.TranslationMatrix(c.Position)
}

func (c *Camera2D) TranslateBy(dx, dy float32) {
	c.Position[0] += -dx
	c.Position[1] += -dy
	c.Position[2] = 0
	c.ViewMatrix = hmath.TranslationMatrix(c.Position)
}

func (c *Camera2D) Projection(canvasWidth int, canvasHeight int) hmath.Matrix4 {
	w := float32(canvasWidth)
	h := float32(canvasHeight)
	hh := h / 2
	wh := w / 2
	scale := 1 / c.Zoom

	return hmath.Ortho(
		-wh*scale,
		wh*scale,
		-hh*scale,
		hh*scale,
		-100, 100,
	)
}
