package glu

import "github.com/go-gl/mathgl/mgl32"

type Vertex [3]float32

type Matrix struct {
	Raw mgl32.Mat4
}

func IdentityMatrix() Matrix {
	return Matrix{mgl32.Ident4()}
}

func TranslationMatrix(v Vertex) Matrix {
	return Matrix{mgl32.Translate3D(v[0], v[1], v[2])}
}

func (m Matrix) Floats() [16]float32 {
	return [16]float32{
		m.Raw[0], m.Raw[1], m.Raw[2], m.Raw[3],
		m.Raw[4], m.Raw[5], m.Raw[6], m.Raw[7],
		m.Raw[8], m.Raw[9], m.Raw[10], m.Raw[11],
		m.Raw[12], m.Raw[13], m.Raw[14], m.Raw[15],
	}
}

func Ortho2D(left, right, bottom, top float32) Matrix {
	return Matrix{mgl32.Ortho(left, right, bottom, top, -1, 1)}
}
