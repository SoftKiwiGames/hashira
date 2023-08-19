package glu

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
)

type Image struct {
	Image image.Image

	Width  int
	Height int
}

func ParseHEXColor(hex string) [4]float32 {
	var c color.RGBA
	c.A = 0xff
	_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	if err == nil {
		return [4]float32{float32(c.R) / 255, float32(c.G) / 255, float32(c.B) / 255, float32(c.A) / 255}
	}

	return [4]float32{1, 0, 1, 1}
}

func LoadImagePNG(url string) (*Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code for image response: %v", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading image: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %v", err)
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	return &Image{
		Image:  img,
		Width:  width,
		Height: height,
	}, nil
}

func (img *Image) ByteSize() int {
	return img.Width * img.Height * 4
}

func (img *Image) Pixels() []byte {
	pixels := make([]byte, img.ByteSize())
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			r, g, b, a := img.Image.At(x, y).RGBA()
			pixels[(y*img.Width+x)*4+0] = byte(r >> 8)
			pixels[(y*img.Width+x)*4+1] = byte(g >> 8)
			pixels[(y*img.Width+x)*4+2] = byte(b >> 8)
			pixels[(y*img.Width+x)*4+3] = byte(a >> 8)
		}
	}
	return pixels
}
