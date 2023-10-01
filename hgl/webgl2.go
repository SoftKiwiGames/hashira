package hgl

// Goal:
//
// Provide a thin wrapper around the WebGL2 API
// that is using native go types.
// Avoid js.Value completely on inputs and outputs.
//

import (
	"fmt"
	"syscall/js"

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
type Buffer js.Value
type Framebuffer js.Value
type FramebufferTarget int
type FramebufferStatus int
type FramebufferAttachment int
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

type WebGL struct {
	gl     hjs.WebGL2RenderingContext
	Canvas hjs.Canvas

	Texture2D TextureType
	RGBA      PixelFormat
	RGB       PixelFormat

	StaticDraw            BufferUsage
	DynamicDraw           BufferUsage
	ArrayBuffer           BufferType
	ElementArrayBuffer    BufferType
	UniformBuffer         BufferType
	VertexArrayObjectNone VertexArrayObject

	FramebufferNone        Framebuffer
	Framebuffer            FramebufferTarget
	DrawFramebuffer        FramebufferTarget
	ReadFramebuffer        FramebufferTarget
	ColorAttachment0       FramebufferAttachment // up to 15 in WebGL2
	DepthAttachment        FramebufferAttachment
	StencilAttachment      FramebufferAttachment
	DepthStencilAttachment FramebufferAttachment

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

	// at least 8 is guaranteed
	Texture0         TextureUnit
	Texture1         TextureUnit
	Texture2         TextureUnit
	Texture3         TextureUnit
	Texture4         TextureUnit
	Texture5         TextureUnit
	Texture6         TextureUnit
	Texture7         TextureUnit
	TextureNone      Texture
	TextureMinFilter TextureParameterName
	TextureMagFilter TextureParameterName
	TextureWrapS     TextureParameterName
	TextureWrapT     TextureParameterName
	Nearest          TextureParameter
	Linear           TextureParameter
	ClampToEdge      TextureParameter

	CompileStatus                          ShaderParameter
	LinkStatus                             ProgramParameter
	ValidateStatus                         ProgramParameter
	FramebufferComplete                    FramebufferStatus
	FramebufferIncompleteAttachment        FramebufferStatus
	FramebufferIncompleteMissingAttachment FramebufferStatus
	FramebufferIncompleteDimensions        FramebufferStatus
	FramebufferUnsupported                 FramebufferStatus
	FramebufferIncompleteMultisample       FramebufferStatus

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
		RGB:       PixelFormat(gl.GetInt("RGB")),

		StaticDraw:         BufferUsage(gl.GetInt("STATIC_DRAW")),
		DynamicDraw:        BufferUsage(gl.GetInt("DYNAMIC_DRAW")),
		ArrayBuffer:        BufferType(gl.GetInt("ARRAY_BUFFER")),
		ElementArrayBuffer: BufferType(gl.GetInt("ELEMENT_ARRAY_BUFFER")),
		UniformBuffer:      BufferType(gl.GetInt("UNIFORM_BUFFER")),

		Framebuffer:            FramebufferTarget(gl.GetInt("FRAMEBUFFER")),
		DrawFramebuffer:        FramebufferTarget(gl.GetInt("DRAW_FRAMEBUFFER")),
		ReadFramebuffer:        FramebufferTarget(gl.GetInt("READ_FRAMEBUFFER")),
		ColorAttachment0:       FramebufferAttachment(gl.GetInt("COLOR_ATTACHMENT0")),
		DepthAttachment:        FramebufferAttachment(gl.GetInt("DEPTH_ATTACHMENT")),
		StencilAttachment:      FramebufferAttachment(gl.GetInt("STENCIL_ATTACHMENT")),
		DepthStencilAttachment: FramebufferAttachment(gl.GetInt("DEPTH_STENCIL_ATTACHMENT")),

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
		Texture1:         TextureUnit(gl.GetInt("TEXTURE1")),
		Texture2:         TextureUnit(gl.GetInt("TEXTURE2")),
		Texture3:         TextureUnit(gl.GetInt("TEXTURE3")),
		Texture4:         TextureUnit(gl.GetInt("TEXTURE4")),
		Texture5:         TextureUnit(gl.GetInt("TEXTURE5")),
		Texture6:         TextureUnit(gl.GetInt("TEXTURE6")),
		Texture7:         TextureUnit(gl.GetInt("TEXTURE7")),
		TextureMinFilter: TextureParameterName(gl.GetInt("TEXTURE_MIN_FILTER")),
		TextureMagFilter: TextureParameterName(gl.GetInt("TEXTURE_MAG_FILTER")),
		TextureWrapS:     TextureParameterName(gl.GetInt("TEXTURE_WRAP_S")),
		TextureWrapT:     TextureParameterName(gl.GetInt("TEXTURE_WRAP_T")),
		Nearest:          TextureParameter(gl.GetInt("NEAREST")),
		Linear:           TextureParameter(gl.GetInt("LINEAR")),
		ClampToEdge:      TextureParameter(gl.GetInt("CLAMP_TO_EDGE")),

		CompileStatus:                          ShaderParameter(gl.GetInt("COMPILE_STATUS")),
		LinkStatus:                             ProgramParameter(gl.GetInt("LINK_STATUS")),
		ValidateStatus:                         ProgramParameter(gl.GetInt("VALIDATE_STATUS")),
		FramebufferComplete:                    FramebufferStatus(gl.GetInt("FRAMEBUFFER_COMPLETE")),
		FramebufferIncompleteAttachment:        FramebufferStatus(gl.GetInt("FRAMEBUFFER_INCOMPLETE_ATTACHMENT")),
		FramebufferIncompleteMissingAttachment: FramebufferStatus(gl.GetInt("FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT")),
		FramebufferIncompleteDimensions:        FramebufferStatus(gl.GetInt("FRAMEBUFFER_INCOMPLETE_DIMENSIONS")),
		FramebufferUnsupported:                 FramebufferStatus(gl.GetInt("FRAMEBUFFER_UNSUPPORTED")),
		FramebufferIncompleteMultisample:       FramebufferStatus(gl.GetInt("FRAMEBUFFER_INCOMPLETE_MULTISAMPLE")),

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

func (w *WebGL) CreateFramebuffer() Framebuffer {
	return Framebuffer(w.gl.Call("createFramebuffer"))
}

func (w *WebGL) BindFramebuffer(target FramebufferTarget, framebuffer Framebuffer) {
	w.gl.Call("bindFramebuffer", int(target), js.Value(framebuffer))
}

func (w *WebGL) CheckFramebufferStatus(target FramebufferTarget) FramebufferStatus {
	return FramebufferStatus(w.gl.Call("checkFramebufferStatus", int(target)).Int())
}

func (w *WebGL) FramebufferTexture2D(target FramebufferTarget, attachment FramebufferAttachment, textarget TextureType, texture Texture) {
	w.gl.Call("framebufferTexture2D", int(target), int(attachment), int(textarget), js.Value(*texture), 0)
}

func (w *WebGL) CreateVertexArray() VertexArrayObject {
	return VertexArrayObject(w.gl.Call("createVertexArray"))
}

func (w *WebGL) CreateTexture() Texture {
	tex := w.gl.Call("createTexture")
	return Texture(&tex)
}

func (w *WebGL) DeleteTexture(texture Texture) {
	w.gl.Call("deleteTexture", js.Value(*texture))
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

func (w *WebGL) Uniform1Int(location Location, value int) {
	w.gl.Call("uniform1i", js.Value(location), value)
}

func (w *WebGL) TexImage2DRGBA(width int, height int, data []byte) {
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

func (w *WebGL) TexImage2DRGB(width int, height int, data []byte) {
	pixels := hjs.NewUInt8Array(data)
	w.gl.Call(
		"texImage2D",
		int(w.Texture2D),
		0, /*mipmap level*/
		int(w.RGB),
		width,
		height,
		0, /*border*/
		int(w.RGB),
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

func (w *WebGL) DrawElements(mode DrawMode, count int, offset int) {
	w.gl.Call("drawElements", int(mode), count, int(w.UnsignedInt), int(offset))
}
