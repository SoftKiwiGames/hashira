package hashira

type Tileset struct {
	Name   string
	Width  int
	Height int
}

func (t *Tileset) TextureUV(tile int, tileWidth int, tileHeight int) (float32, float32, float32, float32) {
	if t == nil {
		return 0, 0, 1, 1
	}

	tilesPerRow := t.Width / tileWidth
	rowX := tile % tilesPerRow
	rowY := tile / tilesPerRow
	w := float32(t.Width)
	h := float32(t.Height)

	u0 := float32(rowX*tileWidth) / w
	v0 := float32(rowY*tileHeight) / h
	u1 := float32((rowX+1)*tileWidth-1) / w
	v1 := float32((rowY+1)*tileHeight-1) / h

	return u0, v0, u1, v1
}
