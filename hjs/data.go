package hjs

import "syscall/js"

var Nil = js.ValueOf(nil)

func NewUInt8Array(array []byte) js.Value {
	jsArray := js.Global().Get("Uint8Array").New(len(array))
	js.CopyBytesToJS(jsArray, array)
	return jsArray
}
