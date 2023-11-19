package evaluator

import (
	"donkey/object"
	"fmt"
	"io/ioutil"
	"net/http"
)

var builtins = map[string]*object.Builtin{
	"len":   builtinLen(),
	"fetch": builtinFetch(),
	"print": builtinPrint(),
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

func builtinPrint() *object.Builtin {
	return &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) == 0 {
				return newError("wrong number of arguments. got=%d, want= >0", nil, len(args))
			}

			var out []string
			for _, arg := range args {
				out = append(out, arg.Inspect())
			}
			fmt.Println(out)
			return NULL
		},
	}
}

func builtinFetch() *object.Builtin {
	return &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", nil, len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				resp, err := http.Get(arg.Value)
				if err != nil {
					return newError("`fetch` request failed, got=%s", nil, err)
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return newError("error reading response body, got=%s", nil, err)
				}

				return &object.String{Value: string(body)}

			default:
				return newError("argument to `fetch` not supported, got=%s", nil, args[0].Type())
			}
		},
	}
}
