package glu

import "syscall/js"

type Game interface {
	Init() error
	Tick(dt float32)
}

func RenderLoop(game Game) {
	if err := game.Init(); err != nil {
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

			// run game tick
			game.Tick(float32(deltaTime))

			// request next frame
			js.Global().Call("requestAnimationFrame", js.FuncOf(callback))
			return nil
		}
		callback(js.ValueOf(nil), []js.Value{js.ValueOf(0)})
	}()
}
