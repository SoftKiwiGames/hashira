package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"syscall/js"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/qbart/wasm-office/internal/glu"
	webgl "github.com/seqsense/webgl-go"
)

const vsSource = `
attribute vec3 position;
attribute vec3 color;
varying vec3 vColor;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

void main(void) {
  gl_Position = projection * view * model * vec4(position, 1.0);
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

type Camera2D struct {
	ViewMatrix glu.Matrix
	Position   glu.Vertex
}

func (c *Camera2D) Translate(x, y float32) {
	c.Position[0] = -x
	c.Position[1] = -y
	c.Position[2] = 0
	c.ViewMatrix = glu.Matrix{mgl32.Translate3D(c.Position[0], c.Position[1], c.Position[2])}
}

type OfficeData struct {
	TileSize  int `json:"tileSize"`
	MapWidth  int `json:"mapWidth"`
	MapHeight int `json:"mapHeight"`
}

type OfficeWidget struct {
	config OfficeData
	fps    int
	dur    float32
	width  int
	height int
	x      float32

	gl            *webgl.WebGL
	locModel      webgl.Location
	locView       webgl.Location
	locProjection webgl.Location
	camera        *Camera2D
	vertices      *glu.VertexBuffer3f
	colors        *glu.VertexBuffer3f
}

func (o *OfficeWidget) Init() error {
	canvas := js.Global().Get("document").Call("getElementById", "glcanvas")
	rawData := canvas.Call("getAttribute", "data-office").String()

	err := json.Unmarshal([]byte(rawData), &o.config)
	if err != nil {
		return err
	}
	fmt.Println("map size:", o.config.MapWidth, "x", o.config.MapHeight)
	fmt.Println("tile size:", o.config.TileSize, "x", o.config.TileSize)

	gl, err := webgl.New(canvas)
	if err != nil {
		return err
	}
	o.gl = gl
	o.width = gl.Canvas.ClientWidth()
	o.height = gl.Canvas.ClientHeight()

	// map size * 6 vertices per tile (we could share vertices between tiles but this is easier)
	n := o.config.MapWidth * o.config.MapHeight * 6
	vertices := glu.NewVertexBuffer3f(n)
	o.vertices = vertices
	colors := glu.NewVertexBuffer3f(n)
	o.colors = colors

	o.camera = &Camera2D{}
	// move to the center of map
	o.camera.Translate(float32(o.config.MapWidth)/2, float32(o.config.MapHeight)/2)

	for i := 0; i < vertices.Len(); i += 6 {
		x := float32(i / 6 % o.config.MapWidth)
		y := float32(i / 6 / o.config.MapWidth)

		// first triangle
		//    2
		//  / |
		// 0--1
		//
		vertices.Set(i+0, x, y, 0)
		vertices.Set(i+1, x+1, y, 0)
		vertices.Set(i+2, x+1, y+1, 0)

		if i == 0 {
			colors.Set(i+0, 1, 1, 1)
			colors.Set(i+1, 0, 0, 0)
			colors.Set(i+2, 1, 0, 1)
		} else {

			colors.Set(i+0, 1, 0, 0)
			colors.Set(i+1, 0, 1, 0)
			colors.Set(i+2, 0, 0, 1)
		}

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
		return err
	}

	if fs, err = initFragmentShader(gl, fsSource); err != nil {
		return err
	}

	program, err := linkShaders(gl, nil, vs, fs)
	if err != nil {
		return err
	}

	gl.UseProgram(program)
	// orthographic projection with origin at center
	matProjection := glu.Matrix{mgl32.Ortho2D(
		-float32(o.width/o.config.TileSize)/2,
		float32(o.width/o.config.TileSize)/2,
		-float32(o.height/o.config.TileSize)/2,
		float32(o.height/o.config.TileSize)/2,
	)}
	matModel := glu.IdentityMatrix()

	o.locModel = gl.GetUniformLocation(program, "model")
	o.locView = gl.GetUniformLocation(program, "view")
	o.locProjection = gl.GetUniformLocation(program, "projection")

	gl.UniformMatrix4fv(o.locModel, false, matModel)
	gl.UniformMatrix4fv(o.locView, false, o.camera.ViewMatrix)
	gl.UniformMatrix4fv(o.locProjection, false, matProjection)

	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	position := gl.GetAttribLocation(program, "position")
	gl.VertexAttribPointer(position, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(position)

	gl.BindBuffer(gl.ARRAY_BUFFER, colorBuffer)
	color := gl.GetAttribLocation(program, "color")
	gl.VertexAttribPointer(color, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(color)

	return nil
}

func (o *OfficeWidget) Tick(dt float32) {
	gl := o.gl

	gl.UniformMatrix4fv(o.locView, false, o.camera.ViewMatrix)
	gl.ClearColor(0.5, 0.5, 0.5, 0.9)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.Enable(gl.DEPTH_TEST)
	gl.Viewport(0, 0, o.width, o.height)
	gl.DrawArrays(gl.TRIANGLES, 0, o.vertices.Len())

	// Debug

	o.dur += dt
	o.fps++
	if o.dur > 1 {
		fmt.Println("fps", o.fps)
		o.fps = 0
		o.dur = 0.0
	}
}

func main() {
	game := &OfficeWidget{}
	glu.RenderLoop(game)
	select {}
}
