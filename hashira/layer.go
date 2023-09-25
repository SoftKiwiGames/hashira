package hashira

type Layer struct {
	Z    float32
	Data [][]int
}

func (l *Layer) Tile(x int, y int) int {
	return l.Data[y][x]
}

func (l *Layer) SetTile(x int, y int, tile int) {
	l.Data[y][x] = tile
}
