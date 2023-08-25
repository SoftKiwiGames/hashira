package main

import (
	"encoding/json"
	"errors"
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
	Size       int         `json:"size"`
	Animations []Animation `json:"animations"`
}

type Animation struct {
	Frames []int   `json:"frames"`
	Delay  float32 `json:"delay"`
}

type Layer struct {
	ID   string  `json:"id"`
	Data [][]int `json:"data"`
	Z    float32 `json:"z"`
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

func (o *Map) Tile(layer int, x int, y int) int {
	return o.Layers[layer].Data[y][x]
}

type Tileset struct {
	TileSize      int
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

func (a *AnimatedTile) Tile() int {
	return a.Animation.Frames[a.FrameIndex]
}

type Mesh struct {
	VertexData *glu.VertexBuffer3f
	SubMeshes  []*SubMesh

	MapWidth  int
	MapHeight int
}

type SubMesh struct {
	Model   glu.Matrix4
	UVs     *glu.VertexBuffer2f
	Tileset *Tileset
	Mesh    *Mesh
}

func TileUV(tile int, tileSize int, tilesetWidth int, tilesetHeight int) (float32, float32, float32, float32) {
	tilesPerRow := tilesetWidth / tileSize
	rowX := tile % tilesPerRow
	rowY := tile / tilesPerRow

	u := float32(rowX*tileSize) / float32(tilesetHeight)
	u2 := float32((rowX+1)*tileSize) / float32(tilesetWidth)
	v := float32(rowY*tileSize) / float32(tilesetHeight)
	v2 := float32((rowY+1)*tileSize) / float32(tilesetHeight)

	return u, v, u2, v2
}

func (s *SubMesh) SetTileAt(x, y int, tile int) {
	i := (y*int(s.Mesh.MapWidth) + x) * 6

	u, v, u2, v2 := TileUV(tile, s.Tileset.TileSize, s.Tileset.TextureWidth, s.Tileset.TextureHeight)

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

	program       glu.Program
	vao           glu.VertexArrayObject
	GL            *glu.WebGL
	GLX           *glu.WebGLExtended
	locModel      glu.Location
	locView       glu.Location
	locProjection glu.Location
	locTileset    glu.Location
	texTileset    glu.Texture

	vertexBuffer glu.Buffer
	uvBuffer     glu.Buffer

	camera          *Camera2D
	matModel        glu.Matrix4
	matProjection   glu.Matrix4
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
	o.GLX = gl.Extended()
	glx := o.GLX
	o.canvasWidth, o.canvasHeight = gl.CanvasSize()

	o.camera = &Camera2D{}
	// move to the center of map
	cx, cy := o.config.Center()
	o.camera.Translate(cx, cy)

	// map size * 6 vertices per tile (we could share vertices between tiles but this is easier)
	n := o.config.MapWidth() * o.config.MapHeight() * 6

	mesh := &Mesh{
		VertexData: glu.NewVertexBuffer3f(n),
		SubMeshes:  make([]*SubMesh, len(o.config.Layers)),
		MapWidth:   o.config.MapWidth(),
		MapHeight:  o.config.MapHeight(),
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

	tileSize := o.config.Tiles.Size
	tileset := &Tileset{
		TileSize:      tileSize,
		TextureWidth:  img.Width,
		TextureHeight: img.Height,
	}

	for i := range o.config.Layers {
		mesh.SubMeshes[i] = &SubMesh{
			Model:   glu.TranslationMatrix(glu.Vertex{0, 0, o.config.Layers[i].Z}),
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
				tile := o.config.Tile(i, mx, o.config.MapHeight()-my-1)
				mesh.SubMeshes[i].SetTileAt(mx, my, tile)
			}
		}
	}

	program, err := glx.CreateDefaultProgram(vsSource, fsSource)
	if err != nil {
		return err
	}
	o.program = program

	gl.UseProgram(program)
	// orthographic projection with origin at center
	o.matProjection = glu.Ortho(
		-float32((o.canvasWidth)/o.config.Tiles.Size)/(2*o.config.Zoom),
		float32((o.canvasWidth)/o.config.Tiles.Size)/(2*o.config.Zoom),
		-float32((o.canvasHeight)/o.config.Tiles.Size)/(2*o.config.Zoom),
		float32((o.canvasHeight)/o.config.Tiles.Size)/(2*o.config.Zoom),
		-10, 10,
	)
	o.matModel = glu.IdentityMatrix()
	o.locModel = gl.GetUniformLocation(program, "model")
	o.locView = gl.GetUniformLocation(program, "view")
	o.locProjection = gl.GetUniformLocation(program, "projection")
	o.locTileset = gl.GetUniformLocation(program, "tileset")

	// VAO
	o.vao = gl.CreateVertexArray()
	o.vertexBuffer = gl.CreateBuffer()
	o.uvBuffer = gl.CreateBuffer()
	gl.BindVertexArray(o.vao)
	glx.AssignAttribToBuffer(program, "position", o.vertexBuffer, gl.Float, 3)
	glx.AssignAttribToBuffer(program, "uv", o.uvBuffer, gl.Float, 2)

	o.texTileset = glx.CreateDefaultTexture(img)

	return nil
}

func (o *Widget) Tick(dt float32) {
	gl := o.GL
	glx := o.GLX

	gl.Enable(gl.DepthTest)
	glx.EnableTransparency()

	gl.Viewport(0, 0, o.canvasWidth, o.canvasHeight)
	gl.ClearColor(o.backgroundColor[0], o.backgroundColor[1], o.backgroundColor[2], o.backgroundColor[3])
	gl.Clear(gl.ColorBufferBit | gl.DepthBufferBit)

	gl.UseProgram(o.program)
	gl.UniformMatrix4(o.locModel, o.matModel)
	gl.UniformMatrix4(o.locView, o.camera.ViewMatrix)
	gl.UniformMatrix4(o.locProjection, o.matProjection)

	glx.BindTexture2D(o.texTileset)

	gl.BindVertexArray(o.vao)

	gl.BindBuffer(gl.ArrayBuffer, o.vertexBuffer)
	glx.BufferDataF(gl.ArrayBuffer, o.mesh.VertexData.Data(), gl.DynamicDraw)

	gl.BindBuffer(gl.ArrayBuffer, o.uvBuffer)
	for _, animatedTile := range o.animatedTiles {
		if animatedTile.Update(dt) {
			o.mesh.SubMeshes[animatedTile.Layer].SetTileAt(animatedTile.X, animatedTile.Y, animatedTile.Tile())
		}
	}

	for _, subMesh := range o.mesh.SubMeshes {
		gl.UniformMatrix4(o.locModel, subMesh.Model)
		glx.BufferDataF(gl.ArrayBuffer, subMesh.UVs.Data(), gl.DynamicDraw)
		glx.DrawTriangles(0, o.mesh.VertexData.Len())
	}
	glx.UnbindAll()
}

func main() {
	game := &Widget{}
	glu.RenderLoop(game)
	select {}
}
