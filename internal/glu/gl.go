package glu

import (
	"syscall/js"

	webgl "github.com/seqsense/webgl-go"
)

type WebGL struct {
	gl    *webgl.WebGL
	rawGL js.Value

	DYNAMIC_DRAW webgl.BufferUsage
}

func NewWebGL(canvas js.Value) (*WebGL, error) {
	gl, err := webgl.New(canvas)
	if err != nil {
		return nil, err
	}
	rawGL := canvas.Call("getContext", "webgl2")

	GL_DYNAMIC_DRAW := webgl.BufferUsage(rawGL.Get("DYNAMIC_DRAW").Int())

	return &WebGL{
		DYNAMIC_DRAW: GL_DYNAMIC_DRAW,
		gl:           gl,
		rawGL:        rawGL,
	}, nil
}

func (w *WebGL) GL() *webgl.WebGL {
	return w.gl
}

func (w *WebGL) TexImage2D(width int, height int, data []byte) {
	pixels := NewUInt8Array(data)
	w.rawGL.Call(
		"texImage2D",
		int(w.gl.TEXTURE_2D),
		0, /*mipmap level*/
		int(w.gl.RGBA),
		width,
		height,
		0, /*border*/
		int(w.gl.RGBA),
		int(w.gl.UNSIGNED_BYTE),
		pixels,
	)

}

func (w *WebGL) BufferData(target webgl.BufferType, data []float32, usage webgl.BufferUsage) {
	w.gl.BufferData(target, webgl.Float32ArrayBuffer(data), usage)
}
