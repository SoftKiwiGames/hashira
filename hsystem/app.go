package hsystem

import (
	"fmt"

	"github.com/qbart/hashira/hashira"
	"github.com/qbart/hashira/hgl"
	"github.com/qbart/hashira/hjs"
	"github.com/qbart/hashira/hmath"
)

type App interface {
	Init() error
	Tick(dt float32)
}

type DefaultApp struct {
	Canvas   hjs.Canvas
	Commands *Commands

	canvasWidth  int
	canvasHeight int

	GL  *hgl.WebGL
	GLX *hgl.WebGLExtended

	hasTileset bool

	programQuad, program           hgl.Program
	locModel                       hgl.Location
	locView                        hgl.Location
	locProjection                  hgl.Location
	locQuad, locTileset            hgl.Location
	vaoQuad, vao                   hgl.VertexArrayObject
	vertexBufferQuad, vertexBuffer hgl.Buffer
	uvBufferQuad, uvBuffer         hgl.Buffer
	fbo                            hgl.Framebuffer
	quadTexture                    hgl.Texture

	world           *hashira.World
	camera          *hashira.Camera2D
	matModel        hmath.Matrix4
	matProjection   hmath.Matrix4
	backgroundColor hgl.Color
}

func (app *DefaultApp) Init() error {
	gl, err := hgl.NewWebGL(app.Canvas)
	if err != nil {
		return err
	}
	glx := gl.Extended()
	app.GL = gl
	app.GLX = glx
	app.canvasWidth, app.canvasHeight = app.Canvas.GetClientWidth(), app.Canvas.GetClientHeight()

	app.world = hashira.New()
	app.backgroundColor = hgl.Color{0, 0, 0, 1}

	app.camera = &hashira.Camera2D{Zoom: 1}

	// shader tileset
	program, err := glx.CreateDefaultProgram(VertexShaderSource, FragmentShaderSource)
	if err != nil {
		return err
	}
	app.program = program
	gl.UseProgram(program)
	app.matProjection = hmath.IdentityMatrix()
	app.matModel = hmath.IdentityMatrix()
	app.locModel = gl.GetUniformLocation(program, "model")
	app.locView = gl.GetUniformLocation(program, "view")
	app.locProjection = gl.GetUniformLocation(program, "projection")
	app.locTileset = gl.GetUniformLocation(program, "tileset")
	// VAO tileset
	app.vao = gl.CreateVertexArray()
	app.vertexBuffer = gl.CreateBuffer()
	app.uvBuffer = gl.CreateBuffer()
	gl.BindVertexArray(app.vao)
	glx.AssignAttribToBuffer(program, "position", app.vertexBuffer, gl.Float, 3)
	glx.AssignAttribToBuffer(program, "uv", app.uvBuffer, gl.Float, 2)

	// shader fbo
	programQuad, err := glx.CreateDefaultProgram(QuadVertexShaderSource, QuadFragmentShaderSource)
	if err != nil {
		return err
	}
	app.programQuad = programQuad
	gl.UseProgram(programQuad)
	app.locQuad = gl.GetUniformLocation(programQuad, "quad")
	// VAO fbo
	app.vaoQuad = gl.CreateVertexArray()
	app.vertexBufferQuad = gl.CreateBuffer()
	app.uvBufferQuad = gl.CreateBuffer()
	gl.BindVertexArray(app.vaoQuad)
	glx.AssignAttribToBuffer(programQuad, "position", app.vertexBufferQuad, gl.Float, 3)
	glx.AssignAttribToBuffer(programQuad, "uv", app.uvBufferQuad, gl.Float, 2)
	gl.BindBuffer(gl.ArrayBuffer, app.vertexBufferQuad)
	glx.BufferDataF(gl.ArrayBuffer, []float32{
		-1.0, -1.0, 0.0, // Bottom-left corner
		1.0, -1.0, 0.0, // Bottom-right corner
		-1.0, 1.0, 0.0, // Top-left corner
		1.0, -1.0, 0.0, // Bottom-right corner
		1.0, 1.0, 0.0, // Top-right corner
		-1.0, 1.0, 0.0, // Top-left corner
	}, gl.StaticDraw)
	gl.BindBuffer(gl.ArrayBuffer, app.uvBufferQuad)
	glx.BufferDataF(gl.ArrayBuffer, []float32{
		0.0, 0.0, // Bottom-left corner
		1.0, 0.0, // Bottom-right corner
		0.0, 1.0, // Top-left corner
		1.0, 0.0, // Bottom-right corner
		1.0, 1.0, // Top-right corner
		0.0, 1.0, // Top-left corner
	}, gl.StaticDraw)

	// fbo
	app.fbo = gl.CreateFramebuffer()
	app.quadTexture = glx.CreateEmptyTextureRGBA(app.canvasWidth, app.canvasHeight)
	gl.BindFramebuffer(gl.Framebuffer, app.fbo)
	gl.FramebufferTexture2D(gl.Framebuffer, gl.ColorAttachment0, gl.Texture2D, app.quadTexture)

	if status := gl.CheckFramebufferStatus(gl.Framebuffer); status != gl.FramebufferComplete {
		return fmt.Errorf("framebuffer error: %v", glx.FramebufferStatusError(status))
	}
	gl.BindFramebuffer(gl.Framebuffer, gl.FramebufferNone)

	return nil
}

func (app *DefaultApp) Tick(dt float32) {
	gl := app.GL
	glx := app.GLX

	app.world.Sync()

	// for _, animatedTile := range o.animatedTiles {
	// 	if animatedTile.Update(dt) {
	// 		SetTileAt(o.world.Maps.Get("main"), mesh.SubMeshes[animatedTile.Layer], &o.tileset, animatedTile.X, animatedTile.Y, animatedTile.Tile())
	// 	}
	// }

	// first pass - render to framebuffer
	gl.BindFramebuffer(gl.Framebuffer, app.fbo)

	gl.Enable(gl.DepthTest)
	glx.EnableTransparency()

	gl.Viewport(0, 0, app.canvasWidth, app.canvasHeight)
	glx.ClearColor(app.backgroundColor)
	gl.Clear(gl.ColorBufferBit | gl.DepthBufferBit)

	gl.UseProgram(app.program)
	gl.UniformMatrix4(app.locModel, app.matModel)
	gl.UniformMatrix4(app.locView, app.camera.ViewMatrix)
	gl.UniformMatrix4(app.locProjection, app.matProjection)

	if app.world.Resources.HasTileset() {
		glx.BindTexture2D(app.world.Resources.Texture)
	}

	gl.BindVertexArray(app.vao)

	app.world.Maps.ForEach(func(name string, m *hashira.Map) {
		gl.BindBuffer(gl.ArrayBuffer, app.vertexBuffer)
		glx.BufferDataF(gl.ArrayBuffer, m.Mesh.VertexData.Data(), gl.DynamicDraw)

		gl.BindBuffer(gl.ArrayBuffer, app.uvBuffer)
		for _, subMesh := range m.Mesh.SubMeshes {
			gl.UniformMatrix4(app.locModel, subMesh.Model)
			glx.BufferDataF(gl.ArrayBuffer, subMesh.UVs.Data(), gl.DynamicDraw)
			glx.DrawTriangles(0, m.Mesh.VertexData.Len())
		}
	})
	glx.UnbindAll()

	// second pass - render framebuffer to canvas
	gl.Disable(gl.DepthTest)
	gl.BindFramebuffer(gl.Framebuffer, gl.FramebufferNone)
	glx.BindTexture2D(app.quadTexture)
	gl.Viewport(0, 0, app.canvasWidth, app.canvasHeight)
	glx.ClearColor(app.backgroundColor)
	gl.Clear(gl.ColorBufferBit)

	gl.UseProgram(app.programQuad)
	gl.BindVertexArray(app.vaoQuad)
	glx.DrawTriangles(0, 6)

	if app.Commands.HasEvents() {
		event := app.Commands.PeekEvent()
		app.handleEvent(event)
	}
}

func (app *DefaultApp) handleEvent(event *Event) {
	switch event.Type {
	// resources
	case "resources.LoadTileset":
		data := event.Data.GetBytes("data")
		img, err := app.world.Resources.LoadTileset(data)
		if err != nil {
			fmt.Println("Error loading tileset: ", err)
			return
		}
		app.world.Resources.Texture = app.GLX.CreateDefaultTextureRGBA(img)
		app.world.Resync()

	// world
	case "world.SetBackground":
		color := event.Data.GetString("color")
		app.backgroundColor = hgl.ParseHEXColor(color)

	case "world.AddMap":
		name := event.Data.GetString("name")
		width := event.Data.GetInt("width")
		height := event.Data.GetInt("height")
		tileWidth := event.Data.GetInt("tileWidth")
		tileHeight := event.Data.GetInt("tileHeight")
		app.world.AddMap(name, width, height, tileWidth, tileHeight)
		app.matProjection = app.camera.Projection(app.canvasWidth, app.canvasHeight)

	case "world.AddLayer":
		mapName := event.Data.GetString("map")
		name := event.Data.GetString("name")
		z := event.Data.GetFloat32("z")
		app.world.AddLayer(mapName, name, z)

	case "world.AddLayerData":
		mapName := event.Data.GetString("map")
		layerName := event.Data.GetString("layer")
		data := event.Data.GetIntArrayOfIntArray("data")
		app.world.AddLayerData(mapName, layerName, data)

	case "world.SetTile":
		mapName := event.Data.GetString("map")
		layerName := event.Data.GetString("layer")
		x := event.Data.GetInt("x")
		y := event.Data.GetInt("y")
		tile := event.Data.GetInt("tile")
		app.world.SetTile(mapName, layerName, x, y, tile)

	case "camera.Translate":
		x := event.Data.GetFloat32("x")
		y := event.Data.GetFloat32("y")
		app.camera.Translate(x, y)

	case "camera.TranslateBy":
		dx := event.Data.GetFloat32("x")
		dy := event.Data.GetFloat32("y")
		app.camera.TranslateBy(dx, dy)

	case "camera.Zoom":
		zoom := event.Data.GetFloat32("zoom")
		app.camera.SetZoom(zoom)
		app.matProjection = app.camera.Projection(app.canvasWidth, app.canvasHeight)

	case "camera.ZoomBy":
		zoom := event.Data.GetFloat32("delta")
		app.camera.ZoomBy(zoom)
		app.matProjection = app.camera.Projection(app.canvasWidth, app.canvasHeight)

	case "camera.TranslateToMapCenter":
		name := event.Data.GetString("map")
		m := app.world.Maps.Get(name)
		cx, cy := m.Center()
		cx *= float32(m.TileWidth)
		cy *= float32(m.TileHeight)
		app.camera.Translate(cx, cy)

	default:
		fmt.Println("Unknown event: ", event.Type)
	}
}
