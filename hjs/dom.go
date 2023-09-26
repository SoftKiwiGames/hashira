package hjs

import "syscall/js"

type Node js.Value

func GetElementByID(id string) Node {
	return Node(js.Global().Get("document").Call("getElementById", id))
}

func (n Node) GetAttribute(name string) js.Value {
	return js.Value(n).Call("getAttribute", name)
}

func (n Node) GetInt(name string) int {
	return js.Value(n).Get(name).Int()
}

func (n Node) SetInt(name string, value int) {
	js.Value(n).Set(name, value)
}

func (n Node) IsNull() bool {
	return js.Value(n).IsNull()
}

func (n Node) JS() js.Value {
	return js.Value(n)
}
