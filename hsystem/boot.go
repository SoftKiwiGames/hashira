package hsystem

import (
	"fmt"
	"syscall/js"

	"github.com/qbart/hashira/hjs"
)

func Init() {
	js.Global().Set("HashiraInitRenderLoop", js.FuncOf(InitRenderLoop))
}

func InitRenderLoop(this js.Value, args []js.Value) any {
	if len(args) != 1 && len(args) != 2 {
		panic("Hashira render loop: expected 1 or 2 arguments - canvasID, {options}")
	}
	canvasID := args[0].String()
	canvas := hjs.Canvas(hjs.GetElementByID(canvasID))
	if canvas.IsNull() {
		return fmt.Errorf("CanvasID: `%s` not found", canvasID)
	}
	resize := hjs.Object(args[1]).GetBool("resize")
	if resize {
		canvas.Resize()
	}

	commands := &Commands{
		Events: make([]*Event, 0, 10),
	}
	js.Global().Set("HashiraSendEvent", js.FuncOf(commands.AddEvent))

	app := &DefaultApp{
		Commands: commands,
		Canvas:   canvas,
	}

	RenderLoop(app)
	return nil
}

func RenderLoop(app App) {
	if err := app.Init(); err != nil {
		panic(err)
	}

	var callback func(this js.Value, args []js.Value) interface{}
	prevTotalDuration := 0.0
	go func() {
		callback = func(this js.Value, args []js.Value) interface{} {
			// calculate delta time
			totalDuration := args[0].Float() / 1000.0
			deltaTime := totalDuration - prevTotalDuration
			prevTotalDuration = totalDuration

			// run single frame
			app.Tick(float32(deltaTime))

			// request next frame
			js.Global().Call("requestAnimationFrame", js.FuncOf(callback))
			return nil
		}
		callback(js.ValueOf(nil), []js.Value{js.ValueOf(0)})
	}()
}
