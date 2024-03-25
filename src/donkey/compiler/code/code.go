package code

import (
	"bytes"
	"fmt"
)

const (
	OpConstant Opcode = iota
	OpPop

	OpJump
	OpJumpNotTruthy

	// Access
	OpIndex
	OpCall
	OpReturn
	OpReturnValue

	OpGetGlobal
	OpSetGlobal
	OpGetLocal
	OpSetLocal

	// Globals
	OpNull
	OpTrue
	OpFalse

	// Prefix operators
	OpMinus
	OpBang

	//Comparison
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpGreaterThanOrEqual

	// Arithmetics
	OpAdd
	OpSubtract
	OpMultiply
	OpDivide

	// Datastructure
	OpArray
	OpHash
)

type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			i++ // bug: Ensure i is incremented to avoid infinite loop
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}
	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operand count for %s\n", def.Name)
}

type Opcode byte

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
	OpPop:      {"OpPop", []int{}},

	OpCall:        {"OpCall", []int{}},
	OpReturn:      {"OpReturn", []int{}},
	OpReturnValue: {"OpReturnValue", []int{}},
	OpIndex:       {"OpIndex", []int{}},

	OpGetGlobal: {"OpGetGlobal", []int{2}},
	OpSetGlobal: {"OpSetGlobal", []int{2}},
	OpGetLocal:  {"OpGetLocal", []int{1}},
	OpSetLocal:  {"OpSetLocal", []int{1}},

	OpJump:          {"OpJump", []int{2}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},

	OpNull:  {"OpNull", []int{}},
	OpTrue:  {"OpTrue", []int{}},
	OpFalse: {"OpFalse", []int{}},

	OpMinus: {"OpMinus", []int{}},
	OpBang:  {"OpBang", []int{}},

	OpEqual:              {"OpEqual", []int{}},
	OpNotEqual:           {"OpNotEqual", []int{}},
	OpGreaterThan:        {"OpGreaterThan", []int{}},
	OpGreaterThanOrEqual: {"OpGreaterThanOrEqual", []int{}},

	OpAdd:      {"OpAdd", []int{}},
	OpSubtract: {"OpSubtract", []int{}},
	OpMultiply: {"OpMultiply", []int{}},
	OpDivide:   {"OpDivide", []int{}},

	OpArray: {"OpArray", []int{2}},
	OpHash:  {"OpHash", []int{2}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}
