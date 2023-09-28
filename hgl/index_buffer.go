package hgl

type IndexBuffer struct {
	data []uint32
}

func NewIndexBuffer(n int) *IndexBuffer {
	return &IndexBuffer{
		data: make([]uint32, n),
	}
}

func (ib *IndexBuffer) Len() int {
	return len(ib.data)
}

func (ib *IndexBuffer) At(i int) uint32 {
	return ib.data[i]
}

func (ib *IndexBuffer) Set(i int, v uint32) {
	ib.data[i] = v
}

func (ib *IndexBuffer) SetTriangle(i int, a, b, c uint32) {
	i *= 3
	ib.data[i+0] = a
	ib.data[i+1] = b
	ib.data[i+2] = c
}

func (ib *IndexBuffer) SetQuad(i int, a, b, c, d uint32) {
	i *= 6

	// first triangle
	// c
	// | \
	// a--b
	ib.data[i+0] = a
	ib.data[i+1] = b
	ib.data[i+2] = c

	// second triangle
	// c--d
	//  \ |
	//    b
	ib.data[i+3] = c
	ib.data[i+4] = b
	ib.data[i+5] = d
}

func (ib *IndexBuffer) Data() []uint32 {
	return ib.data
}
