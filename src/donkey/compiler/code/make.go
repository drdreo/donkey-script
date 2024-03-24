package code

import (
	"donkey/utils"
	"encoding/binary"
	"fmt"
)

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		fmt.Printf("opcode %d undefined", op)
		return []byte{}
	}

	if len(def.OperandWidths) != len(operands) {
		fmt.Printf(utils.Yellow("opcode %s operands mismatch. ", def.Name)+"got=%d, want=%d", len(operands), len(def.OperandWidths))
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			WriteUint16(instruction[offset:], uint16(o))
		}
		offset += width
	}

	return instruction
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}
		offset += width
	}

	return operands, offset
}

func ReadUint16(instructions Instructions) uint16 {
	return binary.BigEndian.Uint16(instructions)
}

func WriteUint16(instructions Instructions, operand uint16) {
	binary.BigEndian.PutUint16(instructions, operand)
}
