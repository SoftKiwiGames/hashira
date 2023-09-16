package hgl

import (
	"fmt"
	"image/color"
)

type Color [4]float32

func ParseHEXColor(hex string) Color {
	var c color.RGBA
	c.A = 0xff
	_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	if err == nil {
		return Color{float32(c.R) / 255, float32(c.G) / 255, float32(c.B) / 255, float32(c.A) / 255}
	}

	return Color{1, 0, 1, 1}
}
