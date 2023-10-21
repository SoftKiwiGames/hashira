package hevents

type CameraTranslated struct {
	X float32 `json:"x,omitempty"`
	Y float32 `json:"y,omitempty"`
}

type CameraTranslatedBy struct {
	X float32 `json:"x,omitempty"`
	Y float32 `json:"y,omitempty"`
}

type CameraZoomed struct {
	Zoom float32 `json:"zoom,omitempty"`
}

type CameraZoomedBy struct {
	Delta float32 `json:"delta,omitempty"`
}

type CameraTranslatedToMapCenter struct {
	Map string `json:"map,omitempty"`
}
