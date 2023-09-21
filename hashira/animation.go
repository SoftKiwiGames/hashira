package hashira

type Animation struct {
	Frames []int   `json:"frames"`
	Delay  float32 `json:"delay"`
}

type AnimatedTile struct {
	*Animation

	Layer      int
	X          int
	Y          int
	FrameIndex int
	Time       float32
}

func (a *AnimatedTile) Update(dt float32) bool {
	a.Time += dt
	if a.Time >= a.Delay {
		a.Time = 0
		a.FrameIndex++
		if a.FrameIndex >= len(a.Animation.Frames) {
			a.FrameIndex = 0
		}
		return true
	}

	return false
}

func (a *AnimatedTile) Tile() int {
	return a.Animation.Frames[a.FrameIndex]
}

// o.animatedTiles = make([]*AnimatedTile, 0)
// for l, layer := range o.config.Layers {
// 	for y, row := range layer.Data {
// 		for x, tile := range row {
// 			for _, animation := range o.config.Tiles.Animations {
// 				if len(animation.Frames) > 1 {
// 					frame := animation.Frames[0]
// 					if tile == frame {
// 						animation := animation
// 						o.animatedTiles = append(o.animatedTiles, &AnimatedTile{
// 							Animation:  &animation,
// 							X:          x,
// 							Y:          o.config.MapHeight() - y - 1,
// 							Layer:      l,
// 							FrameIndex: 0,
// 						})
// 					}
// 				}
// 			}
// 		}
// 	}
// }
