package hjs

import (
	"math"
	"syscall/js"
)

type Canvas Node
type WebGL2RenderingContext js.Value

func (c Canvas) GetWebGL2RenderingContext() WebGL2RenderingContext {
	return WebGL2RenderingContext(js.Value(c).Call("getContext", "webgl2"))
}

func (c Canvas) IsNull() bool {
	return Node(c).IsNull()
}

func (c Canvas) GetAttribute(name string) js.Value {
	return Node(c).GetAttribute(name)
}

func (c Canvas) DevicePixelRatio() float32 {
	return float32(js.Global().Get("devicePixelRatio").Float())
}

func (c Canvas) GetClientWidth() int {
	return int(
		math.Round(
			float64(Node(c).GetInt("clientWidth")) * float64(c.DevicePixelRatio()),
		),
	)
}

func (c Canvas) GetClientHeight() int {
	return int(
		math.Round(
			float64(Node(c).GetInt("clientHeight")) * float64(c.DevicePixelRatio()),
		),
	)
}

func (c Canvas) Resize() {
	width := c.GetClientWidth()
	height := c.GetClientHeight()
	Node(c).SetInt("width", width)
	Node(c).SetInt("height", height)
}

func (gl WebGL2RenderingContext) IsNull() bool {
	return js.Value(gl).IsNull()
}

func (gl WebGL2RenderingContext) GetInt(name string) int {
	return js.Value(gl).Get(name).Int()
}

func (gl WebGL2RenderingContext) Call(method string, args ...any) js.Value {
	return js.Value(gl).Call(method, args...)
}
