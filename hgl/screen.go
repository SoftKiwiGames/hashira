package hgl

import "math"

type Screen struct {
	Width            int
	Height           int
	DevicePixelRatio float32
}

func (s *Screen) Resize(w, h int) {
	s.Width = int(math.Round(float64(w) * float64(s.DevicePixelRatio)))
	s.Height = int(math.Round(float64(h) * float64(s.DevicePixelRatio)))
}
