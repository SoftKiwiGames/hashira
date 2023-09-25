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

	u := float32(rowX*tileWidth) / float32(t.Width)
	u2 := float32((rowX+1)*tileWidth) / float32(t.Width)
	v := float32(rowY*tileHeight) / float32(t.Height)
	v2 := float32((rowY+1)*tileHeight) / float32(t.Height)

	return u, v, u2, v2
}
