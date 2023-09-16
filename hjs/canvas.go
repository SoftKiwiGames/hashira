package hjs

import "syscall/js"

type Canvas Node
type WebGL2RenderingContext js.Value

func (c Canvas) GetWebGL2RenderingContext() WebGL2RenderingContext {
	return WebGL2RenderingContext(js.Value(c).Call("getContext", "webgl2"))
}

func (c Canvas) GetClientWidth() int {
	return js.Value(c).Get("clientWidth").Int()
}

func (c Canvas) GetClientHeight() int {
	return js.Value(c).Get("clientHeight").Int()
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
