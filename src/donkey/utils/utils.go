package utils

import (
	"fmt"

	"donkey/ast"
	"donkey/lexer"
	"donkey/parser"
	"donkey/object"
	"donkey/token"
)

const (
	colorRed    = "\033[31m"
	colorBlue   = "\033[34m"
	colorYellow = "\033[33;1m"
	colorReset  = "\033[0m"
)

const LangName = "donkey"

func Blue(text string, args ...interface{}) string {
	return fmt.Sprintf(colorBlue+text+colorReset, args...)
}

func Red(text string, args ...interface{}) string {
	return fmt.Sprintf(colorRed+text+colorReset, args...)
}

func Yellow(text string, args ...interface{}) string {
	return fmt.Sprintf(colorYellow+text+colorReset, args...)
}

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
