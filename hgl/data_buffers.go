package hgl

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type BufferData interface {
	Bytes() []byte
	Len() int
}

func NewFloat32ArrayBuffer(data []float32) *Float32ArrayBuffer {
	return &Float32ArrayBuffer{
		Data:  data,
		Cache: new(bytes.Buffer),
	}
}

func NewUInt32ArrayBuffer(data []uint32) *UInt32ArrayBuffer {
	return &UInt32ArrayBuffer{
		Data:  data,
		Cache: new(bytes.Buffer),
	}
}

func NewByteArrayBuffer(data []byte) *ByteArrayBuffer {
	buf := ByteArrayBuffer(data)
	return &buf
}

type Float32ArrayBuffer struct {
	Data  []float32
	Cache *bytes.Buffer
}

type UInt32ArrayBuffer struct {
	Data  []uint32
	Cache *bytes.Buffer
}

type ByteArrayBuffer []byte

func (f *Float32ArrayBuffer) Bytes() []byte {
	f.Cache.Reset() // reset does not shrink the buffer so we can reuse it
	for _, x := range f.Data {
		err := binary.Write(f.Cache, binary.LittleEndian, x)
		if err != nil {
			fmt.Println("Float32ArrayBuffer.Bytes error:", err)
			return nil
		}
	}
	return f.Cache.Bytes()
}

func (f *Float32ArrayBuffer) Len() int {
	return len(f.Data)
}

func (u UInt32ArrayBuffer) Bytes() []byte {
	u.Cache.Reset() // reset does not shrink the buffer so we can reuse it
	for _, x := range u.Data {
		err := binary.Write(u.Cache, binary.LittleEndian, x)
		if err != nil {
			fmt.Println("UInt32ArrayBuffer.Bytes error:", err)
			return nil
		}
	}
	return u.Cache.Bytes()
}

func (u UInt32ArrayBuffer) Len() int {
	return len(u.Data)
}

func (b ByteArrayBuffer) Bytes() []byte {
	return b
}

func (b ByteArrayBuffer) Len() int {
	return len(b)
}
