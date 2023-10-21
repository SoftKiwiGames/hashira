package hevents

type MapAdded struct {
	Name       string `json:"name,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	TileWidth  int    `json:"tile_width,omitempty"`
	TileHeight int    `json:"tile_height,omitempty"`
}

type LayerAdded struct {
	Map  string `json:"map,omitempty"`
	Name string
	Z    float32
}

type LayerDataAdded struct {
	Map   string  `json:"map,omitempty"`
	Layer string  `json:"layer,omitempty"`
	Data  [][]int `json:"data,omitempty"`
}

type TileAssigned struct {
	Map   string `json:"map,omitempty"`
	Layer string `json:"layer,omitempty"`
	X     int    `json:"x,omitempty"`
	Y     int    `json:"y,omitempty"`
	Tile  int    `json:"tile,omitempty"`
}
