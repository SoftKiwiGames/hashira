package hgl

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

	"github.com/qbart/hashira/hjs"
	"github.com/qbart/hashira/hmath"
)

type GLenum uint32
type GLboolean bool
type GLbitfield uint32
type GLbyte int8
type GLshort int16
type GLint int32
type GLsizei int32
type GLintptr int64
type GLsizeiptr int64
type GLubyte uint8
type GLushort uint16
type GLuint uint32
type GLfloat float32
type GLclampf float32
type GLint64 int64

type DrawMode int
type TextureType int
type TextureParameterName int
type TextureParameter int
type TextureUnit int
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
type Parameter int

type Float32ArrayBuffer []float32

func (f Float32ArrayBuffer) Bytes() []byte {
	n := 4 * len(f)

	ptr := unsafe.Pointer(&(f[0]))
	pi := (*[1]byte)(ptr)
	buf := (*pi)[:]
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Len = n
	sh.Cap = n

	// buffer := new(bytes.Buffer)
	// for _, x := range f {
	// 	err := binary.Write(buffer, binary.LittleEndian, x)
	// 	if err != nil {
	// 		fmt.Println("Error:", err)
	// 		return nil
	// 	}
	// }

	// return buffer.Bytes()
	return buf
}

type ByteArrayBuffer []byte

func (b ByteArrayBuffer) Bytes() []byte {
	return b
}

type WebGLExtended struct {
	*WebGL
}

type WebGL struct {
	gl     hjs.WebGL2RenderingContext
	Canvas hjs.Canvas

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

	Texture0         TextureUnit
	TextureMinFilter TextureParameterName
	TextureMagFilter TextureParameterName
	TextureWrapS     TextureParameterName
	TextureWrapT     TextureParameterName
	Nearest          TextureParameter
	ClampToEdge      TextureParameter

	CompileStatus  ShaderParameter
	LinkStatus     ProgramParameter
	ValidateStatus ProgramParameter

	MaxCombinedTextureImageUnits Parameter
	MaxTextureImageUnits         Parameter
	MaxVertexTextureImageUnits   Parameter
	MaxTextureSize               Parameter
	MaxRenderbufferSize          Parameter
	MaxVertexAttribs             Parameter
	MaxVaryingVectors            Parameter
	MaxVertexUniformVectors      Parameter
	MaxFragmentUniformVectors    Parameter
}

func NewWebGL(canvas hjs.Canvas) (*WebGL, error) {
	gl := canvas.GetWebGL2RenderingContext()
	if gl.IsNull() {
		return nil, fmt.Errorf("WebGL2 is not supported")
	}

	return &WebGL{
		Canvas: canvas,
		gl:     gl,

		Texture2D: TextureType(gl.GetInt("TEXTURE_2D")),
		RGBA:      PixelFormat(gl.GetInt("RGBA")),

		DynamicDraw:        BufferUsage(gl.GetInt("DYNAMIC_DRAW")),
		ArrayBuffer:        BufferType(gl.GetInt("ARRAY_BUFFER")),
		ElementArrayBuffer: BufferType(gl.GetInt("ELEMENT_ARRAY_BUFFER")),
		UniformBuffer:      BufferType(gl.GetInt("UNIFORM_BUFFER")),

		Float:        Type(gl.GetInt("FLOAT")),
		UnsignedByte: Type(gl.GetInt("UNSIGNED_BYTE")),
		UnsignedInt:  Type(gl.GetInt("UNSIGNED_INT")),

		VertexShader:   ShaderType(gl.GetInt("VERTEX_SHADER")),
		FragmentShader: ShaderType(gl.GetInt("FRAGMENT_SHADER")),

		DepthTest: Capability(gl.GetInt("DEPTH_TEST")),
		Blend:     Capability(gl.GetInt("BLEND")),

		ColorBufferBit:   BufferMask(gl.GetInt("COLOR_BUFFER_BIT")),
		DepthBufferBit:   BufferMask(gl.GetInt("DEPTH_BUFFER_BIT")),
		StencilBufferBit: BufferMask(gl.GetInt("STENCIL_BUFFER_BIT")),

		SrcAlpha:         BlendFactor(gl.GetInt("SRC_ALPHA")),
		OneMinusSrcAlpha: BlendFactor(gl.GetInt("ONE_MINUS_SRC_ALPHA")),

		Triangles: DrawMode(gl.GetInt("TRIANGLES")),

		Texture0:         TextureUnit(gl.GetInt("TEXTURE0")),
		TextureMinFilter: TextureParameterName(gl.GetInt("TEXTURE_MIN_FILTER")),
		TextureMagFilter: TextureParameterName(gl.GetInt("TEXTURE_MAG_FILTER")),
		TextureWrapS:     TextureParameterName(gl.GetInt("TEXTURE_WRAP_S")),
		TextureWrapT:     TextureParameterName(gl.GetInt("TEXTURE_WRAP_T")),
		Nearest:          TextureParameter(gl.GetInt("NEAREST")),
		ClampToEdge:      TextureParameter(gl.GetInt("CLAMP_TO_EDGE")),

		CompileStatus:  ShaderParameter(gl.GetInt("COMPILE_STATUS")),
		LinkStatus:     ProgramParameter(gl.GetInt("LINK_STATUS")),
		ValidateStatus: ProgramParameter(gl.GetInt("VALIDATE_STATUS")),

		MaxCombinedTextureImageUnits: Parameter(gl.GetInt("MAX_COMBINED_TEXTURE_IMAGE_UNITS")),
		MaxTextureImageUnits:         Parameter(gl.GetInt("MAX_TEXTURE_IMAGE_UNITS")),
		MaxVertexTextureImageUnits:   Parameter(gl.GetInt("MAX_VERTEX_TEXTURE_IMAGE_UNITS")),
		MaxTextureSize:               Parameter(gl.GetInt("MAX_TEXTURE_SIZE")),
		MaxRenderbufferSize:          Parameter(gl.GetInt("MAX_RENDERBUFFER_SIZE")),
		MaxVertexAttribs:             Parameter(gl.GetInt("MAX_VERTEX_ATTRIBS")),
		MaxVaryingVectors:            Parameter(gl.GetInt("MAX_VARYING_VECTORS")),
		MaxVertexUniformVectors:      Parameter(gl.GetInt("MAX_VERTEX_UNIFORM_VECTORS")),
		MaxFragmentUniformVectors:    Parameter(gl.GetInt("MAX_FRAGMENT_UNIFORM_VECTORS")),
	}, nil
}

func (w *WebGL) Extended() *WebGLExtended {
	return &WebGLExtended{w}
}

func (w *WebGLExtended) CreateDefaultTexture(img *Image) Texture {
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

func (w *WebGLExtended) CreateDefaultProgram(vertexShaderSourceCode string, fragmentShaderSourceCode string) (Program, error) {
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
	w.ValidateProgram(program)
	if !w.getProgramParameter(program, w.ValidateStatus).Bool() {
		return Program(js.Null()), errors.New("validation failed: " + w.GetProgramInfoLog(program))
	}

	w.DetachShader(program, vertex)
	w.DeleteShader(vertex)
	w.DetachShader(program, frag)
	w.DeleteShader(frag)

	return program, nil
}

func (w *WebGLExtended) AssignAttribToBuffer(program Program, attrName string, buffer Buffer, typ Type, size int) {
	attrLoc := w.GetAttribLocation(program, attrName)
	w.EnableVertexAttribArray(attrLoc)
	w.BindBuffer(w.ArrayBuffer, buffer)
	w.VertexAttribPointer(attrLoc, size, typ, false, 0, 0)
}

func (w *WebGLExtended) ClearColor(c Color) {
	w.WebGL.ClearColor(c[0], c[1], c[2], c[3])
}

func (w *WebGLExtended) EnableTransparency() {
	w.Enable(w.Blend)
	w.BlendFunc(w.SrcAlpha, w.OneMinusSrcAlpha)
}

func (w *WebGLExtended) BufferDataF(target BufferType, data []float32, usage BufferUsage) {
	w.BufferData(target, Float32ArrayBuffer(data), usage)
}

func (w *WebGLExtended) DrawTriangles(offset int, count int) {
	w.DrawArrays(w.Triangles, offset, count)
}

func (w *WebGLExtended) UnbindAll() {
	w.BindVertexArray(VertexArrayObject{})
	w.BindTexture(w.Texture2D, nil)
}

func (w *WebGLExtended) BindTexture2D(texture Texture) {
	w.BindTexture(w.Texture2D, texture)
}

func (w *WebGLExtended) CreateAndCompileShader(kind ShaderType, sourceCode string) (Shader, error) {
	s := w.CreateShader(kind)
	w.ShaderSource(s, sourceCode)
	w.CompileShader(s)
	if !w.GetShaderCompileStatus(s) {
		compilationLog := w.GetShaderInfoLog(s)
		return Shader(js.Null()), fmt.Errorf("compile failed %v", compilationLog)
	}
	return s, nil
}

func (w *WebGLExtended) GetShaderCompileStatus(shader Shader) bool {
	v := w.getShaderParameter(shader, w.CompileStatus)
	return v.Bool()
}

func (w *WebGL) Enable(capability Capability) {
	w.gl.Call("enable", int(capability))
}

func (w *WebGL) Disable(capability Capability) {
	w.gl.Call("disable", int(capability))
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

func (w *WebGL) ActiveTexture(textureUnit TextureUnit) {
	w.gl.Call("activeTexture", int(textureUnit))
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

func (w *WebGL) DetachShader(program Program, shader Shader) {
	w.gl.Call("detachShader", js.Value(program), js.Value(shader))
}

// inteded for private use to avoid js.Value return
func (w *WebGL) getShaderParameter(shader Shader, param ShaderParameter) js.Value {
	v := w.gl.Call("getShaderParameter", js.Value(shader), int(param))
	return v
}

// inteded for private use to avoid js.Value return
func (w *WebGL) getParameter(param int) js.Value {
	v := w.gl.Call("getParameter", param)
	return v
}

func (w *WebGL) GetInteger(param Parameter) int {
	return w.getParameter(int(param)).Int()
}

func (w *WebGL) GetShaderInfoLog(shader Shader) string {
	return w.gl.Call("getShaderInfoLog", js.Value(shader)).String()
}

// inteded for private use to avoid js.Value return
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

func (w *WebGL) ValidateProgram(program Program) {
	w.gl.Call("validateProgram", js.Value(program))
}

func (w *WebGL) GetUniformLocation(program Program, name string) Location {
	return Location(w.gl.Call("getUniformLocation", js.Value(program), name))
}

func (w *WebGL) GetAttribLocation(program Program, name string) AttribLocation {
	return AttribLocation(uint32(w.gl.Call("getAttribLocation", js.Value(program), name).Int()))
}

func (w *WebGL) UniformMatrix4(location Location, mat hmath.Matrix4) {
	matJS := mat.JsValue()
	// no transpose by default
	w.gl.Call("uniformMatrix4fv", js.Value(location), false, matJS)
}

func (w *WebGL) TexImage2D(width int, height int, data []byte) {
	pixels := hjs.NewUInt8Array(data)
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
	dataJS := hjs.NewUInt8Array(b)
	js.CopyBytesToJS(dataJS, b)
	w.gl.Call("bufferData", int(target), dataJS, int(usage))
}

func (w *WebGL) DrawArrays(mode DrawMode, first int, count int) {
	w.gl.Call("drawArrays", int(mode), first, count)
}
