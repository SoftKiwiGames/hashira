package hgl

import (
	"errors"
	"fmt"
	"syscall/js"
)

type WebGLExtended struct {
	*WebGL
}

type FBO struct {
	VAO VertexArrayObject
	Framebuffer
	VertexBuffer Buffer
	UVBuffer     Buffer
	Texture      Texture
	Program      Program
	Loc          Location
	Width        int
	Height       int
}

func (w *WebGLExtended) CreateFBORenderTarget(width int, height int) (*FBO, error) {
	// shader fbo
	program, err := w.CreateDefaultProgram(QuadVertexShaderSource, QuadFragmentShaderSource)
	if err != nil {
		return nil, err
	}
	w.UseProgram(program)
	loc := w.GetUniformLocation(program, "quad")

	// VAO fbo
	vao := w.CreateVertexArray()
	vertexBuffer := w.CreateBuffer()
	uvBuffer := w.CreateBuffer()
	w.BindVertexArray(vao)
	w.AssignAttribToBuffer(program, "position", vertexBuffer, w.Float, 3)
	w.AssignAttribToBuffer(program, "uv", uvBuffer, w.Float, 2)
	w.BindBuffer(w.ArrayBuffer, vertexBuffer)
	w.BufferDataF(w.ArrayBuffer, NewFloat32ArrayBuffer(QuadVertices), w.StaticDraw)
	w.BindBuffer(w.ArrayBuffer, uvBuffer)
	w.BufferDataF(w.ArrayBuffer, NewFloat32ArrayBuffer(QuadUV), w.StaticDraw)

	// fbo
	fb := w.CreateFramebuffer()
	tex := w.CreateEmptyTextureRGBA(width, height)
	w.BindFramebuffer(w.Framebuffer, fb)
	w.FramebufferTexture2D(w.Framebuffer, w.ColorAttachment0, w.Texture2D, tex)

	if status := w.CheckFramebufferStatus(w.Framebuffer); status != w.FramebufferComplete {
		return nil, fmt.Errorf("framebuffer error: %v", w.FramebufferStatusError(status))
	}
	w.BindFramebuffer(w.Framebuffer, w.FramebufferNone)

	return &FBO{
		Framebuffer:  fb,
		Texture:      tex,
		VAO:          vao,
		VertexBuffer: vertexBuffer,
		UVBuffer:     uvBuffer,
		Program:      program,
		Loc:          loc,
		Width:        width,
		Height:       height,
	}, nil
}

func (fbo *FBO) Resize(w *WebGLExtended, screen Screen) {
	fbo.Width = screen.Width
	fbo.Height = screen.Height

	w.DeleteTexture(fbo.Texture)
	fbo.Texture = w.CreateEmptyTextureRGBA(fbo.Width, fbo.Height)

	w.BindFramebuffer(w.Framebuffer, fbo.Framebuffer)
	w.FramebufferTexture2D(w.Framebuffer, w.ColorAttachment0, w.Texture2D, fbo.Texture)
	w.BindFramebuffer(w.Framebuffer, w.FramebufferNone)
}

func (fbo *FBO) Draw(w *WebGLExtended) {
	w.Disable(w.DepthTest)
	w.ActiveTexture(w.Texture0)
	w.BindTexture2D(fbo.Texture)
	w.Viewport(0, 0, fbo.Width, fbo.Height)
	w.Clear(w.ColorBufferBit)

	w.UseProgram(fbo.Program)
	w.BindVertexArray(fbo.VAO)
	w.DrawTriangles(0, 6)
}

func (w *WebGLExtended) CreateDefaultTextureRGBA(img *Image) Texture {
	texture := w.CreateTexture()
	w.BindTexture(w.Texture2D, texture)
	w.TexParameteri(w.Texture2D, w.TextureWrapS, w.ClampToEdge)
	w.TexParameteri(w.Texture2D, w.TextureWrapT, w.ClampToEdge)
	w.TexParameteri(w.Texture2D, w.TextureMagFilter, w.Nearest)
	w.TexParameteri(w.Texture2D, w.TextureMinFilter, w.Nearest)
	w.TexImage2DRGBA(img.Width, img.Height, img.Pixels())
	w.BindTexture2D(nil)
	return texture
}

func (w *WebGLExtended) CreateEmptyTextureRGBA(width int, height int) Texture {
	texture := w.CreateTexture()
	w.BindTexture(w.Texture2D, texture)
	w.TexParameteri(w.Texture2D, w.TextureWrapS, w.ClampToEdge)
	w.TexParameteri(w.Texture2D, w.TextureWrapT, w.ClampToEdge)
	w.TexParameteri(w.Texture2D, w.TextureMagFilter, w.Nearest)
	w.TexParameteri(w.Texture2D, w.TextureMinFilter, w.Nearest)
	w.TexImage2DRGBA(width, height, make([]byte, width*height*4))
	w.BindTexture2D(nil)
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

func (w *WebGLExtended) BufferDataF(target BufferType, data *Float32ArrayBuffer, usage BufferUsage) {
	w.BufferData(target, data, usage)
}

func (w *WebGLExtended) BufferDataU(target BufferType, data *UInt32ArrayBuffer, usage BufferUsage) {
	w.BufferData(target, data, usage)
}

func (w *WebGLExtended) DrawTriangles(offset int, count int) {
	w.DrawArrays(w.Triangles, offset, count)
}

func (w *WebGLExtended) DrawIndexedTriangles(count int, offset int) {
	w.DrawElements(w.Triangles, count, offset)
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

func (w *WebGLExtended) FramebufferStatusError(status FramebufferStatus) error {
	switch status {
	case w.FramebufferIncompleteAttachment:
		return fmt.Errorf("incomplete attachment")
	case w.FramebufferIncompleteMissingAttachment:
		return fmt.Errorf("incomplete missing attachment")
	case w.FramebufferIncompleteDimensions:
		return fmt.Errorf("incomplete dimensions")
	case w.FramebufferUnsupported:
		return fmt.Errorf("unsupported")
	case w.FramebufferIncompleteMultisample:
		return fmt.Errorf("incomplete multisample")
	default:
		return fmt.Errorf("unknown")
	}
}
