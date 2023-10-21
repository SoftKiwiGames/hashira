package hsystem

import (
	"fmt"

	"github.com/qbart/hashira/hashira"
	"github.com/qbart/hashira/hgl"
	"github.com/qbart/hashira/hjs"
	"github.com/qbart/hashira/hmath"
	"github.com/qbart/hashira/hsystem/hevents"
)

type App interface {
	Init() error
	Tick(dt float32)
}

type DefaultApp struct {
	Canvas   hjs.Canvas
	Commands *Commands

	GL  *hgl.WebGL
	GLX *hgl.WebGLExtended

	screen        *hgl.Screen
	program       hgl.Program
	locModel      hgl.Location
	locView       hgl.Location
	locProjection hgl.Location
	locTileset    hgl.Location
	vao           hgl.VertexArrayObject
	vertexBuffer  hgl.Buffer

	uvBuffer hgl.Buffer
	fbo      *hgl.FBO

	world           *hashira.World
	camera          *hashira.Camera2D
	matModel        hmath.Matrix4
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

	app.world = hashira.New()
	app.backgroundColor = hgl.Color{1, 1, 1, 1}

	app.screen = &hgl.Screen{
		Width:            app.Canvas.GetClientWidthDPR(),
		Height:           app.Canvas.GetClientHeightDPR(),
		DevicePixelRatio: app.Canvas.DevicePixelRatio(),
	}
	app.camera = &hashira.Camera2D{Zoom: 1}

	// shader tileset
	program, err := glx.CreateDefaultProgram(hgl.VertexShaderSource, hgl.FragmentShaderSource)
	if err != nil {
		return err
	}
	app.program = program
	gl.UseProgram(program)
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

	// fbo
	fbo, err := glx.CreateFBORenderTarget(app.screen.Width, app.screen.Height)
	if err != nil {
		return err
	}
	app.fbo = fbo

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
	gl.BindFramebuffer(gl.Framebuffer, app.fbo.Framebuffer)

	gl.Enable(gl.DepthTest)
	glx.EnableTransparency()

	camProjection := app.camera.Projection(app.screen)
	gl.Viewport(0, 0, app.screen.Width, app.screen.Height)
	glx.ClearColor(app.backgroundColor)
	gl.Clear(gl.ColorBufferBit | gl.DepthBufferBit)

	gl.UseProgram(app.program)
	gl.UniformMatrix4(app.locModel, app.matModel)
	gl.UniformMatrix4(app.locView, app.camera.ViewMatrix)
	gl.UniformMatrix4(app.locProjection, camProjection)
	gl.Uniform1Int(app.locTileset, 1)

	if app.world.Resources.HasTileset() {
		gl.ActiveTexture(gl.Texture1)
		glx.BindTexture2D(app.world.Resources.Texture)
	}

	gl.BindVertexArray(app.vao)

	app.world.Maps.ForEach(func(name string, m *hashira.Map) {
		gl.BindBuffer(gl.ArrayBuffer, app.vertexBuffer)
		glx.BufferDataF(gl.ArrayBuffer, m.Mesh.Vertices.Data(), gl.DynamicDraw)

		gl.BindBuffer(gl.ArrayBuffer, app.uvBuffer)
		for _, subMesh := range m.Mesh.SubMeshes {
			gl.UniformMatrix4(app.locModel, subMesh.Model)
			glx.BufferDataF(gl.ArrayBuffer, subMesh.UVs.Data(), gl.DynamicDraw)
			glx.DrawTriangles(0, m.Mesh.Vertices.Len())
		}
	})

	gl.BindTexture(gl.Texture2D, gl.TextureNone)
	gl.BindVertexArray(gl.VertexArrayObjectNone)
	gl.BindFramebuffer(gl.Framebuffer, gl.FramebufferNone)

	// second pass - render framebuffer to canvas
	app.fbo.Draw(glx)

	if app.Commands.HasEvents() {
		event := app.Commands.PeekEvent()
		app.handleEvent(event)
	}
}

func (app *DefaultApp) handleEvent(event *Event) {
	switch event.Type {
	case "TilesetLoaded":
		data := JsonData[hevents.TilesetLoaded](event.JsonData)
		img, err := app.world.Resources.LoadTileset(data.Bytes)
		if err != nil {
			fmt.Println("Error loading tileset: ", err)
			return
		}
		app.world.Resources.Texture = app.GLX.CreateDefaultTextureRGBA(img)
		app.world.Resync()

	case "ScreenResized":
		data := JsonData[hevents.ScreenResized](event.JsonData)
		app.screen.Resize(data.Width, data.Height)
		app.Canvas.Resize()
		app.fbo.Resize(app.GLX, *app.screen)

	case "BackgroundColorSet":
		data := JsonData[hevents.BackgroundColorSet](event.JsonData)
		app.backgroundColor = hgl.ParseHEXColor(data.Color)

	case "MapAdded":
		data := JsonData[hevents.MapAdded](event.JsonData)
		app.world.AddMap(data.Name, data.Width, data.Height, data.TileWidth, data.TileHeight)

	case "LayerAdded":
		data := JsonData[hevents.LayerAdded](event.JsonData)
		app.world.AddLayer(data.Map, data.Name, data.Z)

	case "LayerDataAdded":
		data := JsonData[hevents.LayerDataAdded](event.JsonData)
		app.world.AddLayerData(data.Map, data.Layer, data.Data)

	case "TileAssigned":
		data := JsonData[hevents.TileAssigned](event.JsonData)
		app.world.SetTile(data.Map, data.Layer, data.X, data.Y, data.Tile)

	case "CameraTranslated":
		data := JsonData[hevents.CameraTranslated](event.JsonData)
		app.camera.Translate(data.X, data.Y)

	case "CameraTranslatedBy":
		data := JsonData[hevents.CameraTranslatedBy](event.JsonData)
		app.camera.TranslateBy(data.X, data.Y)

	case "CameraZoomed":
		data := JsonData[hevents.CameraZoomed](event.JsonData)
		app.camera.SetZoom(data.Zoom)

	case "CameraZoomedBy":
		data := JsonData[hevents.CameraZoomedBy](event.JsonData)
		app.camera.ZoomBy(data.Delta)

	case "CameraTranslatedToMapCenter":
		data := JsonData[hevents.CameraTranslatedToMapCenter](event.JsonData)
		m := app.world.Maps.Get(data.Map)
		cx, cy := m.Center()
		cx *= float32(m.TileWidth)
		cy *= float32(m.TileHeight)
		app.camera.Translate(cx, cy)

	default:
		fmt.Println("Unknown event: ", event.Type)
	}
}
