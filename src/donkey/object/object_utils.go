package object

import (
	"donkey/token"
	"fmt"
)

// IsTruthy returns false only if object is NULL or FALSE, everything else is true
func IsTruthy(obj Object) bool {
	switch obj := obj.(type) {
	case *Boolean:
		return obj.Value
	case *Null:
		return false
	default:
		return true
	}
}

func IsError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ERROR_OBJ
	}
	return false
}

func NewError(format string, location *token.TokenLocation, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...), Location: location}
}
