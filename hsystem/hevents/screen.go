package hevents

type ScreenResized struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

type BackgroundColorSet struct {
	Color string `json:"color,omitempty"`
}
