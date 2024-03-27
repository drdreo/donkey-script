package vm

import (
	"donkey/compiler"
	"donkey/compiler/code"
	"donkey/object"
	"fmt"
	"strings"
)

const StackSize = 2048
const GlobalSize = 65536
const MaxFrames = 1024

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type VM struct {
	constants []object.Object
	globals   []object.Object

	stack []object.Object
	sp    int // always points to the next value. Top of stack is stack[sp-1]

	frames      []*Frame
	framesIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants: bytecode.Constants,
		globals:   make([]object.Object, GlobalSize),

		stack: make([]object.Object, StackSize),
		sp:    0,

		frames:      frames,
		framesIndex: 1,
	}
}

func NewWithGlobalState(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().instPointer < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().instPointer++

		ip = vm.currentFrame().instPointer
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpPop:
			vm.pop()

		case code.OpConstant:
			constIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().instPointer += 2

			err := vm.push(vm.constants[constIdx])
			if err != nil {
				return err
			}

		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}

		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().instPointer = pos - 1 // jump to one position before, since the loop will instPointer++

		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().instPointer += 2

			condition := vm.pop()
			if !object.IsTruthy(condition) {
				vm.currentFrame().instPointer = pos - 1
			}

		case code.OpSetGlobal:
			globalIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().instPointer += 2

			vm.globals[globalIdx] = vm.pop()

		case code.OpGetGlobal:
			globalIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().instPointer += 2

			err := vm.push(vm.globals[globalIdx])
			if err != nil {
				return err
			}

		case code.OpSetLocal:
			localIdx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().instPointer += 1

			frame := vm.currentFrame()

			vm.stack[frame.basePointer+int(localIdx)] = vm.pop()

		case code.OpGetLocal:
			localIdx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().instPointer += 1

			frame := vm.currentFrame()

			err := vm.push(vm.stack[frame.basePointer+int(localIdx)])
			if err != nil {
				return err
			}

		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}

		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}

		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan, code.OpGreaterThanOrEqual:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		case code.OpAdd, code.OpSubtract, code.OpMultiply, code.OpDivide:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}

		case code.OpArray:
			length := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().instPointer += 2

			arr := vm.buildArray(vm.sp-length, vm.sp)
			vm.sp -= length

			err := vm.push(arr)
			if err != nil {
				return err
			}

		case code.OpHash:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().instPointer += 2

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements
			err = vm.push(hash)
			if err != nil {
				return err
			}

		case code.OpIndex:
			idx := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, idx)
			if err != nil {
				return err
			}

		case code.OpCall:
			fn, ok := vm.stack[vm.sp-1].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("calling non-function")
			}

			frame := NewFrame(fn, vm.sp)
			vm.pushFrame(frame)
			// ?? verify - vm.sp = frame.basePointer + fn.NumLocals
			// reserve stack space for local function bindings
			vm.sp += fn.NumLocals

		case code.OpReturn:
			// empty return
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(Null)
			if err != nil {
				return err
			}

		case code.OpReturnValue:
			returnValue := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)
			if err != nil {
				return err
			}

		}
	}
	return nil
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: (%s, %s)", leftType, rightType)
	}
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	lVal := left.(*object.Integer).Value
	rVal := right.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = lVal + rVal
	case code.OpSubtract:
		result = lVal - rVal
	case code.OpMultiply:
		result = lVal * rVal
	case code.OpDivide:
		result = lVal / rVal
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	lVal := left.(*object.String).Value
	rVal := right.(*object.String).Value

	var result string

	switch op {
	case code.OpAdd:
		result = lVal + rVal
	case code.OpSubtract:
		result = strings.ReplaceAll(lVal, rVal, "")
	default:
		return fmt.Errorf("unknown string operator: %d", op)
	}

	return vm.push(&object.String{Value: result})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		return vm.executeStringComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(left == right))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(left != right))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue == rightValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue != rightValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	case code.OpGreaterThanOrEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue >= rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeStringComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue == rightValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue != rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func nativeBoolToBooleanObject(b bool) *object.Boolean {
	if b {
		return True
	}
	return False
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) push(obj object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = obj
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.sp-1]
	vm.sp--
	return obj
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()
	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	length := int64(len(arrayObject.Elements))
	maxIdx := length - 1

	// validate out of bounds access
	// [1,2,3][3] || [1,2,3][-4]
	if idx > maxIdx || idx < -length {
		return vm.push(Null)
	}

	access := idx
	// support access from back [1,2,3][-1] --> 3
	if idx < 0 {
		access += maxIdx + 1
	}

	return vm.push(arrayObject.Elements[access])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("invalid hash key: %s", index.Type())
	}
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) buildArray(startIdx int, endIdx int) object.Object {
	elements := make([]object.Object, endIdx-startIdx)

	for i := range endIdx - startIdx {
		elements[i] = vm.stack[i+startIdx]
	}
	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIdx int, endIdx int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := startIdx; i < endIdx; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]
		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}
		hashedPairs[hashKey.HashKey()] = pair
	}
	return &object.Hash{Pairs: hashedPairs}, nil
}
