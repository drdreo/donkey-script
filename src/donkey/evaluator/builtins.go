package evaluator

import (
	"donkey/object"
)

var builtins = map[string]*object.Builtin{
	"len": builtinLen(),
}

func builtinLen() *object.Builtin {
	return &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", nil, len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}

			default:
				return newError("argument to `len` not supported, got=%s", nil, args[0].Type())
			}
		},
	}
}
