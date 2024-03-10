package evaluator

import (
	"donkey/object"
	"donkey/utils"
	"fmt"
	"io/ioutil"
	"net/http"
)

var builtins = map[string]*object.Builtin{
	"len":   builtinLen(),
	"first": builtinFirst(),
	"last":  builtinLast(),
	"rest":  builtinRest(),
	"push":  builtinPush(),
	"print": builtinPrint(),
	"fetch": builtinFetch(),
}

func builtinLen() *object.Builtin {
	return &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return utils.NewError("wrong number of arguments. got=%d, want=1", nil, len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}

			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}

			default:
				return utils.NewError("argument to `len` not supported, got=%s", nil, args[0].Type())
			}
		},
	}
}

func builtinFirst() *object.Builtin {
	return &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return utils.NewError("wrong number of arguments. got=%d, want=1", nil, len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return utils.NewError("argument to `first` must be ARRAY, got=%s", nil, args[0].Type())
			}

			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			return NULL
		},
	}
}

func builtinLast() *object.Builtin {
	return &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return utils.NewError("wrong number of arguments. got=%d, want=1", nil, len(args))
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return utils.NewError("argument to `last` must be ARRAY, got=%s", nil, args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}

			return NULL
		},
	}
}

func builtinRest() *object.Builtin {
	return &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return utils.NewError("wrong number of arguments. got=%d, want=1", nil, len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return utils.NewError("argument to `rest` must be ARRAY, got %s", nil, args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]object.Object, length-1)
				copy(newElements, arr.Elements[1:length])
				return &object.Array{Elements: newElements}
			}

			return NULL
		},
	}
}

func builtinPush() *object.Builtin {
	return &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return utils.NewError("wrong number of arguments. got=%d, want=2", nil, len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return utils.NewError("argument to `push` must be ARRAY, got %s", nil, args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			newElements := make([]object.Object, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]

			return &object.Array{Elements: newElements}
		},
	}
}

func builtinPrint() *object.Builtin {
	return &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return NULL
		},
	}
}

func builtinFetch() *object.Builtin {
	return &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return utils.NewError("wrong number of arguments. got=%d, want=1", nil, len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				resp, err := http.Get(arg.Value) // HTTP GET request
				if err != nil {
					return utils.NewError("`fetch` request failed, got=%s", nil, err)
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return utils.NewError("error reading response body, got=%s", nil, err)
				}

				return &object.String{Value: string(body)}

			default:
				return utils.NewError("argument to `fetch` not supported, got=%s", nil, args[0].Type())
			}
		},
	}
}
