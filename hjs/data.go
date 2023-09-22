package hjs

import (
	"syscall/js"
)

var Nil = js.ValueOf(nil)

func NewUInt8Array(array []byte) js.Value {
	jsArray := js.Global().Get("Uint8Array").New(len(array))
	js.CopyBytesToJS(jsArray, array)
	return jsArray
}

type Object js.Value

func (o Object) GetString(key string) string {
	return js.Value(o).Get(key).String()
}

func (o Object) GetInt(key string) int {
	return js.Value(o).Get(key).Int()
}

func (o Object) GetFloat32(key string) float32 {
	return float32(js.Value(o).Get(key).Float())
}

func (o Object) GetIntArrayOfIntArray(key string) [][]int {
	jsData := js.Value(o).Get(key)
	rows := make([][]int, jsData.Length())

	for y := 0; y < len(rows); y++ {
		jsRowData := jsData.Index(y)
		row := make([]int, jsRowData.Length())
		for x := 0; x < len(row); x++ {
			row[x] = jsRowData.Index(x).Int()
		}
		rows[y] = row
	}

	return rows
}

func (o Object) GetBytes(key string) []byte {
	buffer := js.Value(o).Get(key)
	b := make([]byte, buffer.Length())
	js.CopyBytesToGo(b, buffer)
	return b
}
