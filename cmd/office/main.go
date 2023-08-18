package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"syscall/js"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/qbart/wasm-office/internal/glu"
	webgl "github.com/seqsense/webgl-go"
)

//go:embed assets/tileset.png
var TileSet []byte

const vsSource = `
attribute vec3 position;
attribute vec3 color;
attribute vec2 uv;

varying vec3 vColor;
varying vec2 vUV;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

void main(void) {
  gl_Position = projection * view * model * vec4(position, 1.0);
  vColor = color;
  vUV = uv;
}
`

const fsSource = `
precision mediump float;

varying vec3 vColor;
varying vec2 vUV;
 
uniform sampler2D tileset;

void main(void) {
  gl_FragColor = vec4(vColor, 1.);
  gl_FragColor *= texture2D(tileset, vUV);
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

type Map struct {
	TileSize uint32  `json:"tileSize"`
	Zoom     float32 `json:"zoom"`
	Layers   []Layer `json:"layers"`
}

type Layer struct {
	ID   string     `json:"id"`
	Data [][]uint32 `json:"data"`
}

func (o *Map) Center() (x, y float32) {
	return float32(o.MapWidth()) / 2, float32(o.MapHeight()) / 2
}

func (o *Map) MapWidth() int {
	if len(o.Layers) == 0 {
		return 0
	}
	if len(o.Layers[0].Data) == 0 {
		return 0
	}

	return len(o.Layers[0].Data[0])
}

func (o *Map) MapHeight() int {
	if len(o.Layers) == 0 {
		return 0
	}
	return len(o.Layers[0].Data)
}

func (o *Map) Tile(layer int, x int, y int) uint32 {
	return o.Layers[layer].Data[y][x]
}

type Mesh struct {
	vertices *glu.VertexBuffer3f
	colors   *glu.VertexBuffer3f
	uvs      *glu.VertexBuffer2f
}

type VBO struct {
	Vertex webgl.Buffer
	Color  webgl.Buffer
	UV     webgl.Buffer
}

type Widget struct {
	config       Map
	canvasWidth  int
	canvasHeight int

	jsGL            js.Value
	gl              *webgl.WebGL
	locModel        webgl.Location
	locView         webgl.Location
	locProjection   webgl.Location
	locTileset      webgl.Location
	texTileset      webgl.Texture
	vbo             *VBO
	camera          *Camera2D
	layers          map[string]*Mesh
	tilesetImage    image.Image
	GL_DYNAMIC_DRAW webgl.BufferUsage
}

func (o *Widget) Init() error {
	canvas := js.Global().Get("document").Call("getElementById", "glcanvas")
	if canvas.IsNull() {
		return errors.New("canvas not found")
	}
	rawData := canvas.Call("getAttribute", "data-wasm")
	if rawData.IsNull() {
		return errors.New("[data-wasm] not found")
	}

	err := json.Unmarshal([]byte(rawData.String()), &o.config)
	if err != nil {
		return err
	}
	img, err := png.Decode(bytes.NewReader(TileSet))
	if err != nil {
		return err
	}
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Get the raw pixel data
	var pixels []byte
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels = append(pixels, byte(r>>8), byte(g>>8), byte(b>>8), byte(a>>8))
		}
	}

	gl, err := webgl.New(canvas)
	if err != nil {
		return err
	}
	o.gl = gl
	o.jsGL = canvas.Call("getContext", "webgl2")
	o.GL_DYNAMIC_DRAW = webgl.BufferUsage(o.jsGL.Get("DYNAMIC_DRAW").Int())
	o.canvasWidth = gl.Canvas.ClientWidth()
	o.canvasHeight = gl.Canvas.ClientHeight()

	o.camera = &Camera2D{}
	// move to the center of map
	o.camera.Translate(float32(o.config.MapWidth())/2, float32(o.config.MapHeight())/2)

	o.layers = make(map[string]*Mesh)

	// map size * 6 vertices per tile (we could share vertices between tiles but this is easier)
	n := o.config.MapWidth() * o.config.MapHeight() * 6

	for _, layer := range o.config.Layers {
		mesh := &Mesh{
			vertices: glu.NewVertexBuffer3f(n),
			colors:   glu.NewVertexBuffer3f(n),
			uvs:      glu.NewVertexBuffer2f(n),
		}
		o.layers[layer.ID] = mesh
	}

	tileSize := uint32(o.config.TileSize)
	texW := float32(img.Bounds().Max.X)
	texH := float32(img.Bounds().Max.Y)
	tilesPerRow := uint32(img.Bounds().Max.X) / tileSize
	for zIndex, layerConfig := range o.config.Layers {
		layer := o.layers[layerConfig.ID]
		z := float32(zIndex)

		for my := 0; my < o.config.MapHeight(); my++ {
			for mx := 0; mx < o.config.MapWidth(); mx++ {
				tile := o.config.Tile(zIndex, mx, o.config.MapHeight()-my-1)
				i := (my*o.config.MapWidth() + mx) * 6
				x := float32(i / 6 % o.config.MapWidth())
				y := float32(i / 6 / o.config.MapWidth())

				// first triangle
				//    2
				//  / |
				// 0--1
				//

				rowX := tile % tilesPerRow
				rowY := (tile / tilesPerRow)

				u := float32(rowX*tileSize) / texW
				u2 := (float32((rowX+1)*tileSize) / texW)
				v := (float32(rowY*tileSize) / texH)
				v2 := (float32((rowY+1)*tileSize) / texH)

				layer.vertices.Set(i+0, x, y, z)
				layer.vertices.Set(i+1, x+1, y, z)
				layer.vertices.Set(i+2, x+1, y+1, z)
				layer.uvs.Set(i+0, u, v2)
				layer.uvs.Set(i+1, u2, v2)
				layer.uvs.Set(i+2, u2, v)
				layer.colors.Set(i+0, 1, 1, 1)
				layer.colors.Set(i+1, 1, 1, 1)
				layer.colors.Set(i+2, 1, 1, 1)

				// second triangle
				// 4--3
				// | /
				// 5
				layer.vertices.Set(i+3, x+1, y+1, z)
				layer.vertices.Set(i+4, x, y+1, z)
				layer.vertices.Set(i+5, x, y, z)
				layer.uvs.Set(i+3, u2, v)
				layer.uvs.Set(i+4, u, v)
				layer.uvs.Set(i+5, u, v2)
				layer.colors.Set(i+3, 1, 1, 1)
				layer.colors.Set(i+4, 1, 1, 1)
				layer.colors.Set(i+5, 1, 1, 1)
			}
		}
	}

	vbo := &VBO{
		Vertex: gl.CreateBuffer(),
		Color:  gl.CreateBuffer(),
		UV:     gl.CreateBuffer(),
	}
	o.vbo = vbo

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
	matProjection := glu.Matrix{mgl32.Ortho(
		-float32(uint32(o.canvasWidth)/o.config.TileSize)/(2*o.config.Zoom),
		float32(uint32(o.canvasWidth)/o.config.TileSize)/(2*o.config.Zoom),
		-float32(uint32(o.canvasHeight)/o.config.TileSize)/(2*o.config.Zoom),
		float32(uint32(o.canvasHeight)/o.config.TileSize)/(2*o.config.Zoom),
		float32(-1-len(o.config.Layers)),
		float32(1+len(o.config.Layers)),
	)}
	matModel := glu.IdentityMatrix()

	o.locModel = gl.GetUniformLocation(program, "model")
	o.locView = gl.GetUniformLocation(program, "view")
	o.locProjection = gl.GetUniformLocation(program, "projection")
	o.locTileset = gl.GetUniformLocation(program, "tileset")

	gl.UniformMatrix4fv(o.locModel, false, matModel)
	gl.UniformMatrix4fv(o.locView, false, o.camera.ViewMatrix)
	gl.UniformMatrix4fv(o.locProjection, false, matProjection)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.Vertex)
	positionLoc := gl.GetAttribLocation(program, "position")
	gl.VertexAttribPointer(positionLoc, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(positionLoc)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.Color)
	colorLoc := gl.GetAttribLocation(program, "color")
	gl.VertexAttribPointer(colorLoc, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(colorLoc)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo.UV)
	uvLoc := gl.GetAttribLocation(program, "uv")
	gl.VertexAttribPointer(uvLoc, 2, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(uvLoc)

	o.texTileset = gl.CreateTexture()
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, o.texTileset)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)

	jsPixels := js.Global().Get("Uint8ClampedArray").New(len(pixels))
	js.CopyBytesToJS(jsPixels, pixels)
	o.jsGL.Call(
		"texImage2D",
		int(gl.TEXTURE_2D),
		0, /*mipmap level*/
		int(gl.RGBA),
		img.Bounds().Max.X,
		img.Bounds().Max.Y,
		0, /*border*/
		int(gl.RGBA),
		int(gl.UNSIGNED_BYTE),
		jsPixels,
	)
	gl.BindTexture(gl.TEXTURE_2D, nil)

	return nil
}

func (o *Widget) Tick(dt float32) {
	gl := o.gl

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.Viewport(0, 0, o.canvasWidth, o.canvasHeight)
	gl.ClearColor(0.5, 0.5, 0.5, 0.9)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UniformMatrix4fv(o.locView, false, o.camera.ViewMatrix)
	gl.BindTexture(gl.TEXTURE_2D, o.texTileset)
	for _, layer := range o.config.Layers {
		mesh := o.layers[layer.ID]

		gl.BindBuffer(gl.ARRAY_BUFFER, o.vbo.Vertex)
		gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(mesh.vertices.Data()), o.GL_DYNAMIC_DRAW)

		gl.BindBuffer(gl.ARRAY_BUFFER, o.vbo.Color)
		gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(mesh.colors.Data()), o.GL_DYNAMIC_DRAW)

		gl.BindBuffer(gl.ARRAY_BUFFER, o.vbo.UV)
		gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(mesh.uvs.Data()), o.GL_DYNAMIC_DRAW)

		gl.DrawArrays(gl.TRIANGLES, 0, mesh.vertices.Len())
	}
	gl.BindTexture(gl.TEXTURE_2D, nil)
}

func main() {
	game := &Widget{}
	glu.RenderLoop(game)
	select {}
}
