package evaluator

import (
	"donkey/ast"
	"donkey/object"
	"donkey/token"
	"donkey/utils"
	"fmt"
	"strings"
)

// objects for referencing
var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

// TODO: potentially replace passing the location around with a context (containing the location) instead
func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// Statements
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if utils.IsError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if utils.IsError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	// Expressions
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if utils.IsError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right, &node.Token.Location)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if utils.IsError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if utils.IsError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right, &node.Token.Location)

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if utils.IsError(left) {
			return left
		}
		idx := Eval(node.Index, env)
		if utils.IsError(idx) {
			return idx
		}
		return evalIndexExpression(left, idx, &node.Token.Location)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.CallExpression:
		if isQuoteCall(node) {
			return quote(node.Arguments[0], env)
		}

		fn := Eval(node.Function, env)
		if utils.IsError(fn) {
			return fn
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && utils.IsError(args[0]) {
			return args[0]
		}
		res := applyFunction(fn, &node.Token.Location, args)
		if utils.IsError(res) {
			res.(*object.Error).Location = &node.Token.Location
		}
		return res

	// Literals
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && utils.IsError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: env}
	}

	return nil
}

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	if block.Async {
		return evalAsyncBlockStatement(block, env)
	}

	var res object.Object

	for _, stmt := range block.Statements {
		res = Eval(stmt, env)

		if res != nil {
			rt := res.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return res
			}
		}
	}
	return res
}

func evalAsyncBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	go func() {
		var res object.Object
		for _, stmt := range block.Statements {
			res = Eval(stmt, env)

			if res != nil {
				rt := res.Type()
				if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
					fmt.Printf("Async function tried to return. %s TODO message\n", res)
				}
			}
		}

		if res != nil {
			fmt.Println("Async function tried to return via expression. TODO message")
		}
	}()
	return NULL
}

// ____________
//
// eval Expression
// ____________

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, exp := range exps {
		evaled := Eval(exp, env)
		if utils.IsError(evaled) {
			return []object.Object{evaled}
		}
		result = append(result, evaled)
	}

	return result
}

func evalPrefixExpression(operator string, right object.Object, loc *token.TokenLocation) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusOperatorExpression(right, loc)
	default:
		return utils.NewError("unknown operator: %s%s", loc, operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusOperatorExpression(right object.Object, loc *token.TokenLocation) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return utils.NewError("unknown operator: -%s", loc, right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalIntegerInfixExpression(operator string, left, right object.Object, loc *token.TokenLocation) object.Object {
	lVal := left.(*object.Integer).Value
	rVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: lVal + rVal}
	case "-":
		return &object.Integer{Value: lVal - rVal}
	case "*":
		return &object.Integer{Value: lVal * rVal}
	case "/":
		return &object.Integer{Value: lVal / rVal}

	case "<":
		return nativeBoolToBooleanObject(lVal < rVal)
	case "<=":
		return nativeBoolToBooleanObject(lVal <= rVal)
	case ">":
		return nativeBoolToBooleanObject(lVal > rVal)
	case ">=":
		return nativeBoolToBooleanObject(lVal >= rVal)
	case "==":
		return nativeBoolToBooleanObject(lVal == rVal)
	case "!=":
		return nativeBoolToBooleanObject(lVal != rVal)
	default:
		return utils.NewError("unknown operator: %s %s %s", loc, left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object, loc *token.TokenLocation) object.Object {
	lVal := left.(*object.String).Value
	rVal := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: lVal + rVal}
	case "-":
		return &object.String{Value: strings.ReplaceAll(lVal, rVal, "")}
	case "==":
		return nativeBoolToBooleanObject(lVal == rVal)
	case "!=":
		return nativeBoolToBooleanObject(lVal != rVal)
	default:
		return utils.NewError("unknown operator: %s %s %s", loc, left.Type(), operator, right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object, loc *token.TokenLocation) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right, loc)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right, loc)

	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)

	case left.Type() != right.Type():
		return utils.NewError("type mismatch: %s %s %s", loc, left.Type(), operator, right.Type())
	default:
		return utils.NewError("unknown operator: %s %s %s", loc, left.Type(), operator, right.Type())
	}
}

func evalArrayIndexExpression(arr object.Object, idx object.Object, loc *token.TokenLocation) object.Object {
	ao, ok := arr.(*object.Array)
	if !ok {
		return utils.NewError("type mismatch for array index operation. got=%s", loc, arr.Type())
	}

	i := idx.(*object.Integer).Value
	maxIdx := int64(len(ao.Elements) - 1)

	// check upper boundary for positive and negative index
	// [1,2,3][3] || [1,2,3][-4]
	if i > maxIdx || -i > maxIdx+1 {
		return NULL
	}

	access := i
	// support access from back [1,2,3][-1] --> 3
	if i < 0 {
		access += maxIdx + 1
	}
	return ao.Elements[access]
}

func evalHashIndexExpression(hash object.Object, idx object.Object, loc *token.TokenLocation) object.Object {
	ho, ok := hash.(*object.Hash)
	if !ok {
		return utils.NewError("type mismatch for hash index operation. got=%s", loc, hash.Type())

	}

	hashKey, ok := idx.(object.Hashable)
	if !ok {
		return utils.NewError("unusable as hash key: %s", loc, idx.Type())
	}

	hashed := hashKey.HashKey()

	pair, ok := ho.Pairs[hashed]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalIndexExpression(left object.Object, index object.Object, loc *token.TokenLocation) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index, loc)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index, loc)
	default:
		return utils.NewError("index operator not supported: %s", loc, left.Type())
	}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)

	if utils.IsTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	}
	return NULL
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return utils.NewError("identifier not found: "+node.Value, &node.Token.Location)
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)
	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if utils.IsError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return utils.NewError("unusable as hash key: %s", &node.Token.Location, key.Type())
		}

		value := Eval(valueNode, env)
		if utils.IsError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

// ____________
//
// Utility stuff
// ____________

func applyFunction(fn object.Object, loc *token.TokenLocation, args []object.Object) object.Object {
	switch fun := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fun, args)
		evaled := Eval(fun.Body, extendedEnv)
		return unwrapReturnValue(evaled)

	case *object.Builtin:
		return fun.Fn(args...)

	default:
		return utils.NewError("not a function: %s", loc, fn.Type())
	}
}

// copies existing env values over to new one
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}

// we need to unwrap it otherwise a return statement would bubble up and stop the evaluation
func unwrapReturnValue(obj object.Object) object.Object {
	if rV, ok := obj.(*object.ReturnValue); ok {
		return rV.Value
	}
	return obj
}
