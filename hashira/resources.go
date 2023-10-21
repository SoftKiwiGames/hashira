package hashira

import (
	"github.com/qbart/hashira/hgl"
)

type Resources struct {
	Tileset *Tileset
	Image   *hgl.Image
	Texture hgl.Texture
}

func (r *Resources) LoadTileset(data []byte) (*hgl.Image, error) {
	img, err := hgl.LoadImagePNGFromBytes(data)
	if err != nil {
		return nil, err
	}
	r.Image = img
	r.Tileset = &Tileset{
		Name:   "tileset",
		Width:  img.Width,
		Height: img.Height,
	}
	return img, nil
}

func (r *Resources) HasTileset() bool {
	return r.Tileset != nil
}

func (r *Resources) GetTileset() *Tileset {
	return r.Tileset
}
