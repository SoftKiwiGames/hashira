package hjs

import "syscall/js"

type Node js.Value

func GetElementByID(id string) Node {
	return Node(js.Global().Get("document").Call("getElementById", id))
}

func (n Node) GetAttribute(name string) js.Value {
	return js.Value(n).Call("getAttribute", name)
}

func (n Node) IsNull() bool {
	return js.Value(n).IsNull()
}

func (n Node) JS() js.Value {
	return js.Value(n)
}