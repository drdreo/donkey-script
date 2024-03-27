package vm

import (
	"donkey/compiler/code"
	"donkey/object"
)

type Frame struct {
	fn          *object.CompiledFunction
	instPointer int
	basePointer int
}

func NewFrame(fn *object.CompiledFunction, basePointer int) *Frame {
	return &Frame{
		fn:          fn,
		instPointer: -1,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
