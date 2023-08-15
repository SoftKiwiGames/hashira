package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"syscall/js"

	"github.com/go-gl/mathgl/mgl32"
	webgl "github.com/seqsense/webgl-go"
)

const vsSource = `
attribute vec3 position;
attribute vec3 color;
varying vec3 vColor;

uniform mat4 projection;
uniform mat4 model;

void main(void) {
  gl_Position = projection * model * vec4(position, 1.0);
  vColor = color;
}
`

const fsSource = `
precision mediump float;
varying vec3 vColor;
void main(void) {
  gl_FragColor = vec4(vColor, 1.);
}
`

func initVertexShader(gl *webgl.WebGL, src string) (webgl.Shader, error) {
	s := gl.CreateShader(gl.VERTEX_SHADER)
	gl.ShaderSource(s, src)
	gl.CompileShader(s)
	if !gl.GetShaderParameter(s, gl.COMPILE_STATUS).(bool) {
		compilationLog := gl.GetShaderInfoLog(s)
		return webgl.Shader(js.Null()), fmt.Errorf("compile failed (VERTEX_SHADER) %v", compilationLog)
	}
	return s, nil
}

func initFragmentShader(gl *webgl.WebGL, src string) (webgl.Shader, error) {
	s := gl.CreateShader(gl.FRAGMENT_SHADER)
	gl.ShaderSource(s, src)
	gl.CompileShader(s)
	if !gl.GetShaderParameter(s, gl.COMPILE_STATUS).(bool) {
		compilationLog := gl.GetShaderInfoLog(s)
		return webgl.Shader(js.Null()), fmt.Errorf("compile failed (FRAGMENT_SHADER) %v", compilationLog)
	}
	return s, nil
}

func linkShaders(gl *webgl.WebGL, fbVarings []string, shaders ...webgl.Shader) (webgl.Program, error) {
	program := gl.CreateProgram()
	for _, s := range shaders {
		gl.AttachShader(program, s)
	}
	if len(fbVarings) > 0 {
		gl.TransformFeedbackVaryings(program, fbVarings, gl.SEPARATE_ATTRIBS)
	}
	gl.LinkProgram(program)
	if !gl.GetProgramParameter(program, gl.LINK_STATUS).(bool) {
		return webgl.Program(js.Null()), errors.New("link failed: " + gl.GetProgramInfoLog(program))
	}
	return program, nil
}

type Mat4 struct {
	m mgl32.Mat4
}

func (m Mat4) Floats() [16]float32 {
	return [16]float32{
		m.m[0], m.m[1], m.m[2], m.m[3],
		m.m[4], m.m[5], m.m[6], m.m[7],
		m.m[8], m.m[9], m.m[10], m.m[11],
		m.m[12], m.m[13], m.m[14], m.m[15],
	}
}

func NewVertexBuffer3f(n int) *VertexBuffer3f {
	// n * 3 elements (x, y, z)
	return &VertexBuffer3f{
		data: make([]float32, n*3),
	}
}

type VertexBuffer3f struct {
	data []float32
}

func (v *VertexBuffer3f) Len() int {
	return len(v.data) / 3
}

func (v *VertexBuffer3f) At(i int) (x, y, z float32) {
	i *= 3
	return v.data[i], v.data[i+1], v.data[i+2]
}

func (v *VertexBuffer3f) Set(i int, x, y, z float32) {
	i *= 3
	v.data[i] = x
	v.data[i+1] = y
	v.data[i+2] = z
}

func (v *VertexBuffer3f) Data() []float32 {
	return v.data
}

type OfficeData struct {
	TileSize  int `json:"tileSize"`
	MapWidth  int `json:"mapWidth"`
	MapHeight int `json:"mapHeight"`
}

func run() {
	canvas := js.Global().Get("document").Call("getElementById", "glcanvas")
	rawData := canvas.Call("getAttribute", "data-office").String()

	var officeData OfficeData
	err := json.Unmarshal([]byte(rawData), &officeData)
	if err != nil {
		panic(err)
	}

	gl, err := webgl.New(canvas)
	if err != nil {
		panic(err)
	}

	width := gl.Canvas.ClientWidth()
	height := gl.Canvas.ClientHeight()

	// map size * 6 vertices per tile (we could share vertices between tiles but this is easier)
	panX := float32(officeData.MapWidth) / 2
	panY := float32(officeData.MapHeight) / 2
	panX = 0
	panY = 0
	vertices := NewVertexBuffer3f(officeData.MapWidth * officeData.MapHeight * 6)
	colors := NewVertexBuffer3f(officeData.MapWidth * officeData.MapHeight * 6)

	fmt.Println("map size:", officeData.MapWidth, "x", officeData.MapHeight)
	fmt.Println("tile size:", officeData.TileSize, "x", officeData.TileSize)
	for i := 0; i < vertices.Len(); i += 6 {
		x := float32(i/6%officeData.MapWidth) - panX
		y := float32(i/6/officeData.MapWidth) - panY

		// first triangle
		//    2
		//  / |
		// 0--1
		//
		vertices.Set(i+0, x, y, 0)
		vertices.Set(i+1, x+1, y, 0)
		vertices.Set(i+2, x+1, y+1, 0)

		colors.Set(i+0, 1, 0, 0)
		colors.Set(i+1, 0, 1, 0)
		colors.Set(i+2, 0, 0, 1)

		// second triangle
		// 1--0
		// | /
		// 2
		vertices.Set(i+3, x+1, y+1, 0)
		vertices.Set(i+4, x, y+1, 0)
		vertices.Set(i+5, x, y, 0)

		colors.Set(i+3, 1, 0, 0)
		colors.Set(i+4, 0, 1, 0)
		colors.Set(i+5, 0, 0, 1)
	}

	vertexBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(vertices.Data()), gl.STATIC_DRAW)

	colorBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, colorBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(colors.Data()), gl.STATIC_DRAW)

	var vs, fs webgl.Shader
	if vs, err = initVertexShader(gl, vsSource); err != nil {
		panic(err)
	}

	if fs, err = initFragmentShader(gl, fsSource); err != nil {
		panic(err)
	}

	program, err := linkShaders(gl, nil, vs, fs)
	if err != nil {
		panic(err)
	}

	gl.UseProgram(program)
	matProjection := Mat4{mgl32.Ortho2D(0, float32(width/officeData.TileSize), float32(height/officeData.TileSize), 0)}
	matModel := Mat4{mgl32.Ident4()}

	locProjection := gl.GetUniformLocation(program, "projection")
	locModel := gl.GetUniformLocation(program, "model")

	gl.UniformMatrix4fv(locProjection, false, matProjection)
	gl.UniformMatrix4fv(locModel, false, matModel)

	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	position := gl.GetAttribLocation(program, "position")
	gl.VertexAttribPointer(position, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(position)

	gl.BindBuffer(gl.ARRAY_BUFFER, colorBuffer)
	color := gl.GetAttribLocation(program, "color")
	gl.VertexAttribPointer(color, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(color)

	gl.ClearColor(0.5, 0.5, 0.5, 0.9)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.Enable(gl.DEPTH_TEST)
	gl.Viewport(0, 0, width, height)
	gl.DrawArrays(gl.TRIANGLES, 0, vertices.Len())
}

func main() {
	go run()
	select {}
}
