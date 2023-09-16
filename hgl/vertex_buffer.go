package hgl

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

func NewVertexBuffer2f(n int) *VertexBuffer2f {
	// n * 2 elements (x, y)
	return &VertexBuffer2f{
		data: make([]float32, n*2),
	}
}

type VertexBuffer2f struct {
	data []float32
}

func (v *VertexBuffer2f) Len() int {
	return len(v.data) / 2
}

func (v *VertexBuffer2f) At(i int) (x, y float32) {
	i *= 2
	return v.data[i], v.data[i+1]
}

func (v *VertexBuffer2f) Set(i int, x, y float32) {
	i *= 2
	v.data[i] = x
	v.data[i+1] = y
}

func (v *VertexBuffer2f) Data() []float32 {
	return v.data
}
