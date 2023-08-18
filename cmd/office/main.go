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
  gl_FragColor = texture2D(tileset, vUV);
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
	TileSize uint32     `json:"tileSize"`
	Zoom     float32    `json:"mapZoom"`
	Data     [][]uint32 `json:"mapData"`
}

func (o *Map) Center() (x, y float32) {
	return float32(o.MapWidth()) / 2, float32(o.MapHeight()) / 2
}

func (o *Map) MapWidth() int {
	if len(o.Data) == 0 {
		return 0
	}

	return len(o.Data[0])
}

func (o *Map) MapHeight() int {
	return len(o.Data)
}

func (o *Map) Tile(x, y int) uint32 {
	return o.Data[y][x]
}

type OfficeWidget struct {
	config       Map
	canvasWidth  int
	canvasHeight int

	jsGL          js.Value
	gl            *webgl.WebGL
	locModel      webgl.Location
	locView       webgl.Location
	locProjection webgl.Location
	locTileset    webgl.Location
	texTileset    webgl.Texture
	camera        *Camera2D
	vertices      *glu.VertexBuffer3f
	colors        *glu.VertexBuffer3f
	uvs           *glu.VertexBuffer2f
	tilesetImage  image.Image
}

func (o *OfficeWidget) Init() error {
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
	o.canvasWidth = gl.Canvas.ClientWidth()
	o.canvasHeight = gl.Canvas.ClientHeight()

	// map size * 6 vertices per tile (we could share vertices between tiles but this is easier)
	n := o.config.MapWidth() * o.config.MapHeight() * 6
	vertices := glu.NewVertexBuffer3f(n)
	o.vertices = vertices
	colors := glu.NewVertexBuffer3f(n)
	o.colors = colors
	uvs := glu.NewVertexBuffer2f(n)
	o.uvs = uvs

	o.camera = &Camera2D{}
	// move to the center of map
	o.camera.Translate(float32(o.config.MapWidth())/2, float32(o.config.MapHeight())/2)

	tileSize := uint32(o.config.TileSize)
	texW := float32(img.Bounds().Max.X)
	texH := float32(img.Bounds().Max.Y)
	tilesPerRow := uint32(img.Bounds().Max.X) / tileSize
	for my := 0; my < o.config.MapHeight(); my++ {
		for mx := 0; mx < o.config.MapWidth(); mx++ {
			tile := o.config.Tile(mx, o.config.MapHeight()-my-1)
			i := (my*o.config.MapWidth() + mx) * 6
			x := float32(i / 6 % o.config.MapWidth())
			y := float32(i / 6 / o.config.MapWidth())

			// first triangle
			//    2
			//  / |
			// 0--1
			//
			vertices.Set(i+0, x, y, 0)
			vertices.Set(i+1, x+1, y, 0)
			vertices.Set(i+2, x+1, y+1, 0)

			rowX := tile % tilesPerRow
			rowY := (tile / tilesPerRow)

			u := float32(rowX*tileSize) / texW
			u2 := (float32((rowX+1)*tileSize) / texW)
			v := (float32(rowY*tileSize) / texH)
			v2 := (float32((rowY+1)*tileSize) / texH)

			uvs.Set(i+0, u, v2)
			uvs.Set(i+1, u2, v2)
			uvs.Set(i+2, u2, v)
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
			// 4--3
			// | /
			// 5
			vertices.Set(i+3, x+1, y+1, 0)
			vertices.Set(i+4, x, y+1, 0)
			vertices.Set(i+5, x, y, 0)
			uvs.Set(i+3, u2, v)
			uvs.Set(i+4, u, v)
			uvs.Set(i+5, u, v2)
			colors.Set(i+3, 1, 0, 0)
			colors.Set(i+4, 0, 1, 0)
			colors.Set(i+5, 0, 0, 1)
		}
	}

	vertexBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(vertices.Data()), gl.STATIC_DRAW)

	colorBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, colorBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(colors.Data()), gl.STATIC_DRAW)

	uvBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, uvBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(uvs.Data()), gl.STATIC_DRAW)

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
		-float32(uint32(o.canvasWidth)/o.config.TileSize)/(2*o.config.Zoom),
		float32(uint32(o.canvasWidth)/o.config.TileSize)/(2*o.config.Zoom),
		-float32(uint32(o.canvasHeight)/o.config.TileSize)/(2*o.config.Zoom),
		float32(uint32(o.canvasHeight)/o.config.TileSize)/(2*o.config.Zoom),
	)}
	matModel := glu.IdentityMatrix()

	o.locModel = gl.GetUniformLocation(program, "model")
	o.locView = gl.GetUniformLocation(program, "view")
	o.locProjection = gl.GetUniformLocation(program, "projection")
	o.locTileset = gl.GetUniformLocation(program, "tileset")

	gl.UniformMatrix4fv(o.locModel, false, matModel)
	gl.UniformMatrix4fv(o.locView, false, o.camera.ViewMatrix)
	gl.UniformMatrix4fv(o.locProjection, false, matProjection)

	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	positionLoc := gl.GetAttribLocation(program, "position")
	gl.VertexAttribPointer(positionLoc, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(positionLoc)

	gl.BindBuffer(gl.ARRAY_BUFFER, colorBuffer)
	colorLoc := gl.GetAttribLocation(program, "color")
	gl.VertexAttribPointer(colorLoc, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(colorLoc)

	gl.BindBuffer(gl.ARRAY_BUFFER, uvBuffer)
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

func (o *OfficeWidget) Tick(dt float32) {
	gl := o.gl

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.Viewport(0, 0, o.canvasWidth, o.canvasHeight)
	gl.ClearColor(0.5, 0.5, 0.5, 0.9)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UniformMatrix4fv(o.locView, false, o.camera.ViewMatrix)
	gl.BindTexture(gl.TEXTURE_2D, o.texTileset)
	gl.DrawArrays(gl.TRIANGLES, 0, o.vertices.Len())
	gl.BindTexture(gl.TEXTURE_2D, nil)

	// Debug

	// o.dur += dt
	// o.fps++
	// if o.dur > 1 {
	// 	fmt.Println("fps", o.fps)
	// 	o.fps = 0
	// 	o.dur = 0.0
	// }
}

func main() {
	game := &OfficeWidget{}
	glu.RenderLoop(game)
	select {}
}
