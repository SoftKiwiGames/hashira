package glu

// Goal:
//
// Provide a thin wrapper around the WebGL2 API
// that is using native go types.
// Avoid js.Value completely on inputs and outputs.
//

import (
	"errors"
	"fmt"
	"reflect"
	"syscall/js"
	"unsafe"
)

type DrawMode int
type TextureType int
type TextureParameterName int
type TextureParameter int
type Texture *js.Value
type BufferData interface {
	Bytes() []byte
}
type Buffer js.Value
type BufferUsage int
type BufferType int
type PixelFormat int
type ShaderType int
type Capability int
type Type int
type VertexArrayObject js.Value
type BufferMask int
type BlendFactor int
type AttribLocation uint32
type Location js.Value
type Program js.Value
type Shader js.Value
type ShaderParameter int
type ProgramParameter int

type Float32ArrayBuffer []float32

func (f Float32ArrayBuffer) Bytes() []byte {
	n := 4 * len(f)

	ptr := unsafe.Pointer(&(f[0]))
	pi := (*[1]byte)(ptr)
	buf := (*pi)[:]
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Len = n
	sh.Cap = n

	return buf
}

type ByteArrayBuffer []byte

func (b ByteArrayBuffer) Bytes() []byte {
	return b
}

type WebGL struct {
	gl js.Value

	Canvas js.Value

	Texture2D TextureType
	RGBA      PixelFormat

	DynamicDraw        BufferUsage
	ArrayBuffer        BufferType
	ElementArrayBuffer BufferType
	UniformBuffer      BufferType

	Float        Type
	UnsignedByte Type
	UnsignedInt  Type

	VertexShader   ShaderType
	FragmentShader ShaderType

	DepthTest Capability
	Blend     Capability

	ColorBufferBit   BufferMask
	DepthBufferBit   BufferMask
	StencilBufferBit BufferMask

	SrcAlpha         BlendFactor
	OneMinusSrcAlpha BlendFactor

	Triangles DrawMode

	TextureMinFilter TextureParameterName
	TextureMagFilter TextureParameterName
	TextureWrapS     TextureParameterName
	TextureWrapT     TextureParameterName
	Nearest          TextureParameter
	ClampToEdge      TextureParameter

	CompileStatus  ShaderParameter
	LinkStatus     ProgramParameter
	ValidateStatus ProgramParameter
}

func NewWebGL(canvas js.Value) (*WebGL, error) {
	gl := canvas.Call("getContext", "webgl2")
	if gl.IsNull() {
		return nil, fmt.Errorf("WebGL2 is not supported")
	}

	return &WebGL{
		Canvas: canvas,
		gl:     gl,

		Texture2D: TextureType(gl.Get("TEXTURE_2D").Int()),
		RGBA:      PixelFormat(gl.Get("RGBA").Int()),

		DynamicDraw:        BufferUsage(gl.Get("DYNAMIC_DRAW").Int()),
		ArrayBuffer:        BufferType(gl.Get("ARRAY_BUFFER").Int()),
		ElementArrayBuffer: BufferType(gl.Get("ELEMENT_ARRAY_BUFFER").Int()),
		UniformBuffer:      BufferType(gl.Get("UNIFORM_BUFFER").Int()),

		Float:        Type(gl.Get("FLOAT").Int()),
		UnsignedByte: Type(gl.Get("UNSIGNED_BYTE").Int()),
		UnsignedInt:  Type(gl.Get("UNSIGNED_INT").Int()),

		VertexShader:   ShaderType(gl.Get("VERTEX_SHADER").Int()),
		FragmentShader: ShaderType(gl.Get("FRAGMENT_SHADER").Int()),

		DepthTest: Capability(gl.Get("DEPTH_TEST").Int()),
		Blend:     Capability(gl.Get("BLEND").Int()),

		ColorBufferBit:   BufferMask(gl.Get("COLOR_BUFFER_BIT").Int()),
		DepthBufferBit:   BufferMask(gl.Get("DEPTH_BUFFER_BIT").Int()),
		StencilBufferBit: BufferMask(gl.Get("STENCIL_BUFFER_BIT").Int()),

		SrcAlpha:         BlendFactor(gl.Get("SRC_ALPHA").Int()),
		OneMinusSrcAlpha: BlendFactor(gl.Get("ONE_MINUS_SRC_ALPHA").Int()),

		Triangles: DrawMode(gl.Get("TRIANGLES").Int()),

		TextureMinFilter: TextureParameterName(gl.Get("TEXTURE_MIN_FILTER").Int()),
		TextureMagFilter: TextureParameterName(gl.Get("TEXTURE_MAG_FILTER").Int()),
		TextureWrapS:     TextureParameterName(gl.Get("TEXTURE_WRAP_S").Int()),
		TextureWrapT:     TextureParameterName(gl.Get("TEXTURE_WRAP_T").Int()),
		Nearest:          TextureParameter(gl.Get("NEAREST").Int()),
		ClampToEdge:      TextureParameter(gl.Get("CLAMP_TO_EDGE").Int()),

		CompileStatus:  ShaderParameter(gl.Get("COMPILE_STATUS").Int()),
		LinkStatus:     ProgramParameter(gl.Get("LINK_STATUS").Int()),
		ValidateStatus: ProgramParameter(gl.Get("VALIDATE_STATUS").Int()),
	}, nil
}

// extension beyond standard WebGL
func (w *WebGL) CreateDefaultTexture(img *Image) Texture {
	texture := w.CreateTexture()
	w.BindTexture(w.Texture2D, texture)
	w.TexParameteri(w.Texture2D, w.TextureWrapS, w.ClampToEdge)
	w.TexParameteri(w.Texture2D, w.TextureWrapT, w.ClampToEdge)
	w.TexParameteri(w.Texture2D, w.TextureMagFilter, w.Nearest)
	w.TexParameteri(w.Texture2D, w.TextureMinFilter, w.Nearest)
	w.TexImage2D(img.Width, img.Height, img.Pixels())
	w.BindTexture(w.Texture2D, nil)
	return texture
}

// extension beyond standard WebGL
func (w *WebGL) CreateDefaultProgram(vertexShaderSourceCode string, fragmentShaderSourceCode string) (Program, error) {
	vertex, err := w.CreateAndCompileShader(w.VertexShader, vertexShaderSourceCode)
	if err != nil {
		return Program(js.Null()), fmt.Errorf("VERTEX_SHADER: %v", err)
	}
	frag, err := w.CreateAndCompileShader(w.FragmentShader, fragmentShaderSourceCode)
	if err != nil {
		return Program(js.Null()), fmt.Errorf("FRAGMENT_SHADER: %v", err)
	}

	program := w.CreateProgram()
	w.AttachShader(program, vertex)
	w.AttachShader(program, frag)
	w.LinkProgram(program)

	if !w.getProgramParameter(program, w.LinkStatus).Bool() {
		return Program(js.Null()), errors.New("link failed: " + w.GetProgramInfoLog(program))
	}
	w.DeleteShader(vertex)
	w.DeleteShader(frag)

	return program, nil
}

// extension beyond standard WebGL
func (w *WebGL) AssignAttribToBuffer(program Program, attrName string, buffer Buffer, typ Type, size int) {
	attrLoc := w.GetAttribLocation(program, attrName)
	w.EnableVertexAttribArray(attrLoc)
	w.BindBuffer(w.ArrayBuffer, buffer)
	w.VertexAttribPointer(attrLoc, size, typ, false, 0, 0)
}

func (w *WebGL) CanvasSize() (width, height int) {
	return w.Canvas.Get("clientWidth").Int(),
		w.Canvas.Get("clientHeight").Int()
}

func (w *WebGL) Enable(capability Capability) {
	w.gl.Call("enable", int(capability))
}

func (w *WebGL) Disable(capability Capability) {
	w.gl.Call("disable", int(capability))
}

func (w *WebGL) EnableTransparency() {
	w.Enable(w.Blend)
	w.BlendFunc(w.SrcAlpha, w.OneMinusSrcAlpha)
}

func (w *WebGL) BlendFunc(sfactor, dfactor BlendFactor) {
	w.gl.Call("blendFunc", int(sfactor), int(dfactor))
}

func (w *WebGL) ClearColor(r, g, b, a float32) {
	w.gl.Call("clearColor", r, g, b, a)
}

func (w *WebGL) Clear(mask BufferMask) {
	w.gl.Call("clear", int(mask))
}

func (w *WebGL) Viewport(x, y, width, height int) {
	w.gl.Call("viewport", x, y, width, height)
}

func (w *WebGL) CreateVertexArray() VertexArrayObject {
	return VertexArrayObject(w.gl.Call("createVertexArray"))
}

func (gl *WebGL) CreateTexture() Texture {
	tex := gl.gl.Call("createTexture")
	return Texture(&tex)
}

func (w *WebGL) BindVertexArray(vao VertexArrayObject) {
	w.gl.Call("bindVertexArray", js.Value(vao))
}

func (w *WebGL) BindTexture(textureType TextureType, texture Texture) {
	if texture == nil {
		w.gl.Call("bindTexture", int(textureType), nil)
	} else {
		w.gl.Call("bindTexture", int(textureType), js.Value(*texture))
	}
}

func (gl *WebGL) TexParameteri(texType TextureType, name TextureParameterName, param TextureParameter) {
	gl.gl.Call("texParameteri", int(texType), int(name), int(param))
}

func (w *WebGL) CreateProgram() Program {
	return Program(w.gl.Call("createProgram"))
}

func (w *WebGL) UseProgram(program Program) {
	w.gl.Call("useProgram", js.Value(program))
}

func (w *WebGL) CreateShader(shaderType ShaderType) Shader {
	return Shader(w.gl.Call("createShader", int(shaderType)))
}

func (w *WebGL) DeleteShader(shader Shader) {
	w.gl.Call("deleteShader", js.Value(shader))
}

func (w *WebGL) CompileShader(shader Shader) {
	w.gl.Call("compileShader", js.Value(shader))
}

func (w *WebGL) ShaderSource(shader Shader, source string) {
	w.gl.Call("shaderSource", js.Value(shader), source)
}

func (w *WebGL) AttachShader(program Program, shader Shader) {
	w.gl.Call("attachShader", js.Value(program), js.Value(shader))
}

// extension beyond standard WebGL
func (w *WebGL) CreateAndCompileShader(kind ShaderType, sourceCode string) (Shader, error) {
	s := w.CreateShader(kind)
	w.ShaderSource(s, sourceCode)
	w.CompileShader(s)
	if !w.GetShaderCompileStatus(s) {
		compilationLog := w.GetShaderInfoLog(s)
		return Shader(js.Null()), fmt.Errorf("compile failed %v", compilationLog)
	}
	return s, nil
}

// extension beyond standard WebGL
func (w *WebGL) GetShaderCompileStatus(shader Shader) bool {
	v := w.getShaderParameter(shader, w.CompileStatus)
	return v.Bool()
}

// standard WebGL but only inteded for private use
func (w *WebGL) getShaderParameter(shader Shader, param ShaderParameter) js.Value {
	v := w.gl.Call("getShaderParameter", js.Value(shader), int(param))
	return v
}

func (w *WebGL) GetShaderInfoLog(shader Shader) string {
	return w.gl.Call("getShaderInfoLog", js.Value(shader)).String()
}

// standard WebGL but only inteded for private use
func (w *WebGL) getProgramParameter(p Program, param ProgramParameter) js.Value {
	v := w.gl.Call("getProgramParameter", js.Value(p), int(param))
	return v
}

func (w *WebGL) GetProgramInfoLog(p Program) string {
	return w.gl.Call("getProgramInfoLog", js.Value(p)).String()
}

func (w *WebGL) LinkProgram(program Program) {
	w.gl.Call("linkProgram", js.Value(program))
}

func (w *WebGL) GetUniformLocation(program Program, name string) Location {
	return Location(w.gl.Call("getUniformLocation", js.Value(program), name))
}

func (w *WebGL) GetAttribLocation(program Program, name string) AttribLocation {
	return AttribLocation(uint32(w.gl.Call("getAttribLocation", js.Value(program), name).Int()))
}

func (w *WebGL) UniformMatrix4(location Location, mat Matrix4) {
	matJS := mat.JsValue()
	// no transpose by default
	w.gl.Call("uniformMatrix4fv", js.Value(location), false, matJS)
}

func (w *WebGL) TexImage2D(width int, height int, data []byte) {
	pixels := NewUInt8Array(data)
	w.gl.Call(
		"texImage2D",
		int(w.Texture2D),
		0, /*mipmap level*/
		int(w.RGBA),
		width,
		height,
		0, /*border*/
		int(w.RGBA),
		int(w.UnsignedByte),
		pixels,
	)
}

func (w *WebGL) CreateBuffer() Buffer {
	return Buffer(w.gl.Call("createBuffer"))
}

func (w *WebGL) EnableVertexAttribArray(location AttribLocation) {
	w.gl.Call("enableVertexAttribArray", uint32(location))
}

func (w *WebGL) VertexAttribPointer(location AttribLocation, size int, typ Type, normalized bool, stride byte, offset int) {
	w.gl.Call("vertexAttribPointer", uint32(location), size, int(typ), normalized, stride, offset)
}

func (w *WebGL) BindBuffer(target BufferType, buffer Buffer) {
	w.gl.Call("bindBuffer", int(target), js.Value(buffer))
}

func (w *WebGL) BufferData(target BufferType, data BufferData, usage BufferUsage) {
	b := data.Bytes()
	dataJS := NewUInt8Array(b)
	js.CopyBytesToJS(dataJS, b)
	w.gl.Call("bufferData", int(target), dataJS, int(usage))
}

func (w *WebGL) DrawArrays(mode DrawMode, first int, count int) {
	w.gl.Call("drawArrays", int(mode), first, count)
}
