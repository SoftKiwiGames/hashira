package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"syscall/js"

	"github.com/qbart/hashira/internal/glu"
)

const vsSource = `
attribute vec3 position;
attribute vec2 uv;

varying vec2 vUV;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

void main(void) {
  gl_Position = projection * view * model * vec4(position, 1.0);
  vUV = uv;
}
`

const fsSource = `
precision mediump float;

varying vec2 vUV;
 
uniform sampler2D tileset;

void main(void) {
  gl_FragColor = texture2D(tileset, vUV);
}
`

func initVertexShader(gl *glu.WebGL, src string) (glu.Shader, error) {
	s := gl.CreateShader(gl.VertexShader)
	gl.ShaderSource(s, src)
	gl.CompileShader(s)
	if !gl.GetShaderParameter(s, gl.CompileStatus) {
		compilationLog := gl.GetShaderInfoLog(s)
		return glu.Shader(js.Null()), fmt.Errorf("compile failed (VERTEX_SHADER) %v", compilationLog)
	}
	return s, nil
}

func initFragmentShader(gl *glu.WebGL, src string) (glu.Shader, error) {
	s := gl.CreateShader(gl.FragmentShader)
	gl.ShaderSource(s, src)
	gl.CompileShader(s)
	if !gl.GetShaderParameter(s, gl.CompileStatus) {
		compilationLog := gl.GetShaderInfoLog(s)
		return glu.Shader(js.Null()), fmt.Errorf("compile failed (FRAGMENT_SHADER) %v", compilationLog)
	}
	return s, nil
}

func linkShaders(gl *glu.WebGL, shaders ...glu.Shader) (glu.Program, error) {
	program := gl.CreateProgram()
	for _, s := range shaders {
		gl.AttachShader(program, s)
	}
	gl.LinkProgram(program)
	if !gl.GetProgramParameter(program, gl.LinkStatus) {
		return glu.Program(js.Null()), errors.New("link failed: " + gl.GetProgramInfoLog(program))
	}
	// TODO delete shaders
	return program, nil
}

type Camera2D struct {
	ViewMatrix glu.Matrix4
	Position   glu.Vertex
}

func (c *Camera2D) Translate(x, y float32) {
	c.Position[0] = -x
	c.Position[1] = -y
	c.Position[2] = 0
	c.ViewMatrix = glu.TranslationMatrix(c.Position)
}

type Map struct {
	Background string  `json:"background"`
	Tiles      Tiles   `json:"tiles"`
	Zoom       float32 `json:"zoom"`
	Layers     []Layer `json:"layers"`
}
type Tiles struct {
	URL        string      `json:"url"`
	Size       uint32      `json:"size"`
	Animations []Animation `json:"animations"`
}

type Animation struct {
	Frames []uint32 `json:"frames"`
	Delay  float32  `json:"delay"`
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

type Tileset struct {
	TileSize      uint32
	TilesPerRow   uint32
	TextureWidth  int
	TextureHeight int
}

type AnimatedTile struct {
	*Animation

	Layer      int
	X          int
	Y          int
	FrameIndex int
	Time       float32
}

func (a *AnimatedTile) Update(dt float32) bool {
	a.Time += dt
	if a.Time >= a.Delay {
		a.Time = 0
		a.FrameIndex++
		if a.FrameIndex >= len(a.Animation.Frames) {
			a.FrameIndex = 0
		}
		return true
	}

	return false
}

func (a *AnimatedTile) Tile() uint32 {
	return a.Animation.Frames[a.FrameIndex]
}

type Mesh struct {
	VertexBuffer glu.Buffer
	UVBuffer     glu.Buffer

	VertexData *glu.VertexBuffer3f
	SubMeshes  []*SubMesh

	MapWidth  uint32
	MapHeight uint32
}

type SubMesh struct {
	ZIndex  int
	UVs     *glu.VertexBuffer2f
	Tileset *Tileset
	Mesh    *Mesh
}

func (s *SubMesh) SetTileAt(x, y int, tile uint32) {
	i := (y*int(s.Mesh.MapWidth) + x) * 6

	rowX := tile % s.Tileset.TilesPerRow
	rowY := (tile / s.Tileset.TilesPerRow)

	u := float32(rowX*s.Tileset.TileSize) / float32(s.Tileset.TextureWidth)
	u2 := float32((rowX+1)*s.Tileset.TileSize) / float32(s.Tileset.TextureWidth)
	v := float32(rowY*s.Tileset.TileSize) / float32(s.Tileset.TextureHeight)
	v2 := float32((rowY+1)*s.Tileset.TileSize) / float32(s.Tileset.TextureHeight)

	// first triangle
	//    2
	//  / |
	// 0--1
	//
	s.UVs.Set(i+0, u, v2)
	s.UVs.Set(i+1, u2, v2)
	s.UVs.Set(i+2, u2, v)
	// second triangle
	// 4--3
	// | /
	// 5
	s.UVs.Set(i+3, u2, v)
	s.UVs.Set(i+4, u, v)
	s.UVs.Set(i+5, u, v2)
}

type Widget struct {
	config       Map
	canvasWidth  int
	canvasHeight int

	program         glu.Program
	vao             glu.VertexArrayObject
	GL              *glu.WebGL
	locModel        glu.Location
	locView         glu.Location
	locProjection   glu.Location
	locTileset      glu.Location
	texTileset      glu.Texture
	camera          *Camera2D
	mesh            *Mesh
	tilesetImage    image.Image
	backgroundColor [4]float32
	animatedTiles   []*AnimatedTile
}

func (o *Widget) Init() error {
	canvas := js.Global().Get("document").Call("getElementById", "hashira-container")
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
	img, err := glu.LoadImagePNG(o.config.Tiles.URL)
	if err != nil {
		return err
	}
	o.backgroundColor = glu.ParseHEXColor(o.config.Background)

	gl, err := glu.NewWebGL(canvas)
	if err != nil {
		return err
	}
	o.GL = gl
	o.canvasWidth, o.canvasHeight = gl.CanvasSize()

	o.camera = &Camera2D{}
	// move to the center of map
	cx, cy := o.config.Center()
	o.camera.Translate(cx, cy)

	// map size * 6 vertices per tile (we could share vertices between tiles but this is easier)
	n := o.config.MapWidth() * o.config.MapHeight() * 6

	mesh := &Mesh{
		VertexData:   glu.NewVertexBuffer3f(n),
		SubMeshes:    make([]*SubMesh, len(o.config.Layers)),
		VertexBuffer: gl.CreateBuffer(),
		UVBuffer:     gl.CreateBuffer(),
		MapWidth:     uint32(o.config.MapWidth()),
		MapHeight:    uint32(o.config.MapHeight()),
	}
	o.mesh = mesh

	o.animatedTiles = make([]*AnimatedTile, 0)
	for l, layer := range o.config.Layers {
		for y, row := range layer.Data {
			for x, tile := range row {
				for _, animation := range o.config.Tiles.Animations {
					if len(animation.Frames) > 1 {
						frame := animation.Frames[0]
						if tile == frame {
							animation := animation
							o.animatedTiles = append(o.animatedTiles, &AnimatedTile{
								Animation:  &animation,
								X:          x,
								Y:          o.config.MapHeight() - y - 1,
								Layer:      l,
								FrameIndex: 0,
							})
						}
					}
				}
			}
		}
	}

	tileSize := uint32(o.config.Tiles.Size)
	tileset := &Tileset{
		TileSize:      tileSize,
		TilesPerRow:   uint32(img.Width) / tileSize,
		TextureWidth:  img.Width,
		TextureHeight: img.Height,
	}

	for i := range o.config.Layers {
		mesh.SubMeshes[i] = &SubMesh{
			ZIndex:  i,
			UVs:     glu.NewVertexBuffer2f(n),
			Mesh:    mesh,
			Tileset: tileset,
		}
	}

	for my := 0; my < o.config.MapHeight(); my++ {
		for mx := 0; mx < o.config.MapWidth(); mx++ {
			z := float32(0)
			i := (my*o.config.MapWidth() + mx) * 6
			x := float32(i / 6 % o.config.MapWidth())
			y := float32(i / 6 / o.config.MapWidth())

			// first triangle
			//    2
			//  / |
			// 0--1
			//

			mesh.VertexData.Set(i+0, x, y, z)
			mesh.VertexData.Set(i+1, x+1, y, z)
			mesh.VertexData.Set(i+2, x+1, y+1, z)

			// second triangle
			// 4--3
			// | /
			// 5
			mesh.VertexData.Set(i+3, x+1, y+1, z)
			mesh.VertexData.Set(i+4, x, y+1, z)
			mesh.VertexData.Set(i+5, x, y, z)
		}
	}

	for i := range o.config.Layers {
		for my := 0; my < o.config.MapHeight(); my++ {
			for mx := 0; mx < o.config.MapWidth(); mx++ {
				tile := o.config.Tile(mesh.SubMeshes[i].ZIndex, mx, o.config.MapHeight()-my-1)
				mesh.SubMeshes[i].SetTileAt(mx, my, tile)
			}
		}
	}

	var vs, fs glu.Shader
	if vs, err = initVertexShader(gl, vsSource); err != nil {
		return err
	}

	if fs, err = initFragmentShader(gl, fsSource); err != nil {
		return err
	}

	program, err := linkShaders(gl, vs, fs)
	if err != nil {
		return err
	}

	gl.UseProgram(program)
	// orthographic projection with origin at center
	matProjection := glu.Ortho2D(
		-float32(uint32(o.canvasWidth)/o.config.Tiles.Size)/(2*o.config.Zoom),
		float32(uint32(o.canvasWidth)/o.config.Tiles.Size)/(2*o.config.Zoom),
		-float32(uint32(o.canvasHeight)/o.config.Tiles.Size)/(2*o.config.Zoom),
		float32(uint32(o.canvasHeight)/o.config.Tiles.Size)/(2*o.config.Zoom),
	)
	matModel := glu.IdentityMatrix()

	o.locModel = gl.GetUniformLocation(program, "model")
	o.locView = gl.GetUniformLocation(program, "view")
	o.locProjection = gl.GetUniformLocation(program, "projection")
	o.locTileset = gl.GetUniformLocation(program, "tileset")

	gl.UniformMatrix4(o.locModel, matModel)
	gl.UniformMatrix4(o.locView, o.camera.ViewMatrix)
	gl.UniformMatrix4(o.locProjection, matProjection)

	// VAO
	o.vao = gl.CreateVertexArray()
	gl.BindVertexArray(o.vao)

	positionLoc := gl.GetAttribLocation(program, "position")
	gl.EnableVertexAttribArray(positionLoc)
	gl.BindBuffer(gl.ArrayBuffer, mesh.VertexBuffer)
	gl.VertexAttribPointer(positionLoc, 3, gl.Float, false, 0, 0)

	uvLoc := gl.GetAttribLocation(program, "uv")
	gl.EnableVertexAttribArray(uvLoc)
	gl.BindBuffer(gl.ArrayBuffer, mesh.UVBuffer)
	gl.VertexAttribPointer(uvLoc, 2, gl.Float, false, 0, 0)
	gl.BindVertexArray(glu.VertexArrayObject{})
	//

	// TEXTURE
	o.texTileset = gl.CreateTexture()
	gl.BindTexture(gl.Texture2D, o.texTileset)
	gl.TexParameteri(gl.Texture2D, gl.TextureWrapS, gl.ClampToEdge)
	gl.TexParameteri(gl.Texture2D, gl.TextureWrapT, gl.ClampToEdge)
	gl.TexParameteri(gl.Texture2D, gl.TextureMagFilter, gl.Nearest)
	gl.TexParameteri(gl.Texture2D, gl.TextureMinFilter, gl.Nearest)
	gl.TexImage2D(img.Width, img.Height, img.Pixels())
	gl.BindTexture(gl.Texture2D, nil)
	//

	return nil
}

func (o *Widget) Tick(dt float32) {
	gl := o.GL

	gl.Disable(gl.DepthTest)
	gl.Enable(gl.Blend)
	gl.BlendFunc(gl.SrcAlpha, gl.OneMinusSrcAlpha)

	gl.Viewport(0, 0, o.canvasWidth, o.canvasHeight)
	gl.ClearColor(o.backgroundColor[0], o.backgroundColor[1], o.backgroundColor[2], o.backgroundColor[3])
	gl.Clear(int(gl.ColorBufferBit))

	gl.BindVertexArray(o.vao)
	gl.UniformMatrix4(o.locView, o.camera.ViewMatrix)
	gl.BindTexture(gl.Texture2D, o.texTileset)

	gl.BindBuffer(gl.ArrayBuffer, o.mesh.VertexBuffer)
	gl.BufferData(gl.ArrayBuffer, glu.Float32ArrayBuffer(o.mesh.VertexData.Data()), gl.DynamicDraw)

	gl.BindBuffer(gl.ArrayBuffer, o.mesh.UVBuffer)
	for _, animatedTile := range o.animatedTiles {
		if animatedTile.Update(dt) {
			o.mesh.SubMeshes[animatedTile.Layer].SetTileAt(animatedTile.X, animatedTile.Y, animatedTile.Tile())
		}
	}
	for _, subMesh := range o.mesh.SubMeshes {
		gl.BufferData(gl.ArrayBuffer, glu.Float32ArrayBuffer(subMesh.UVs.Data()), gl.DynamicDraw)
		gl.DrawArrays(gl.Triangles, 0, o.mesh.VertexData.Len())
	}
	o.GL.BindVertexArray(glu.VertexArrayObject{})
	gl.BindTexture(gl.Texture2D, nil)
}

func main() {
	game := &Widget{}
	glu.RenderLoop(game)
	select {}
}
