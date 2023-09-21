package hashira

type Layer struct {
	Z    float32
	Data [][]int
}

func (l *Layer) Tile(x int, y int) int {
	return l.Data[y][x]
}
