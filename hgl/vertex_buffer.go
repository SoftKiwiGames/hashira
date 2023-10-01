package hgl

func NewVertexBuffer3f(n int) *VertexBuffer3f {
	// n * 3 elements (x, y, z)
	return &VertexBuffer3f{
		data: NewFloat32ArrayBuffer(make([]float32, n*3)),
	}
}

type VertexBuffer3f struct {
	data *Float32ArrayBuffer
}

func (v *VertexBuffer3f) Len() int {
	return v.data.Len() / 3
}

func (v *VertexBuffer3f) At(i int) (x, y, z float32) {
	i *= 3
	return v.data.Data[i], v.data.Data[i+1], v.data.Data[i+2]
}

func (v *VertexBuffer3f) Set(i int, x, y, z float32) {
	i *= 3
	v.data.Data[i] = x
	v.data.Data[i+1] = y
	v.data.Data[i+2] = z
}

func (v *VertexBuffer3f) Data() *Float32ArrayBuffer {
	return v.data
}

func NewVertexBuffer2f(n int) *VertexBuffer2f {
	// n * 2 elements (x, y)
	return &VertexBuffer2f{
		data: NewFloat32ArrayBuffer(make([]float32, n*2)),
	}
}

type VertexBuffer2f struct {
	data *Float32ArrayBuffer
}

func (v *VertexBuffer2f) Len() int {
	return v.data.Len() / 2
}

func (v *VertexBuffer2f) At(i int) (x, y float32) {
	i *= 2
	return v.data.Data[i], v.data.Data[i+1]
}

func (v *VertexBuffer2f) Set(i int, x, y float32) {
	i *= 2
	v.data.Data[i] = x
	v.data.Data[i+1] = y
}

func (v *VertexBuffer2f) SetQuad(i int, u0, v0, u1, v1 float32) {
	i *= 6

	v.Set(i+0, u0, v1)
	v.Set(i+1, u1, v1)
	v.Set(i+2, u1, v0)

	v.Set(i+3, u1, v0)
	v.Set(i+4, u0, v0)
	v.Set(i+5, u0, v1)
}

func (v *VertexBuffer2f) Data() *Float32ArrayBuffer {
	return v.data
}
