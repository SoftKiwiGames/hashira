package glu

func NewVertexBuffer3f(n int) *VertexBuffer3f {
	// n * 3 elements (x, y, z)
	return &VertexBuffer3f{
		data: make([]float32, n*3),
	}
}

type VertexBuffer3f struct {
	data []float32
}

func (v *VertexBuffer3f) Len() int {
	return len(v.data) / 3
}

func (v *VertexBuffer3f) At(i int) (x, y, z float32) {
	i *= 3
	return v.data[i], v.data[i+1], v.data[i+2]
}

func (v *VertexBuffer3f) Set(i int, x, y, z float32) {
	i *= 3
	v.data[i] = x
	v.data[i+1] = y
	v.data[i+2] = z
}

func (v *VertexBuffer3f) Data() []float32 {
	return v.data
}
