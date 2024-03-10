package utils

import (
	"fmt"

	"donkey/ast"
	"donkey/lexer"
	"donkey/parser"
	"donkey/object"
	"donkey/token"
)

func ParseProgram(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

// IsTruthy returns false only if object is NULL or FALSE, everything else is true
func IsTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func NewError(format string, location *token.TokenLocation, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...), Location: location}
}

func IsError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
