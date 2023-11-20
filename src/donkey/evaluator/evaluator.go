package evaluator

import (
	"donkey/ast"
	"donkey/object"
	"donkey/token"
	"fmt"
	"strings"
)

// objects for referencing
var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

// TODO: potentially replace passing the location around with a context instead
func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	// Statements
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	// Expressions
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right, &node.Token.Location)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right, &node.Token.Location)

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		idx := Eval(node.Index, env)
		if isError(idx) {
			return idx
		}
		return evalIndexExpression(left, idx, &node.Token.Location)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.CallExpression:
		fn := Eval(node.Function, env)
		if isError(fn) {
			return fn
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		res := applyFunction(fn, &node.Token.Location, args)
		if isError(res) {
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
		if len(elements) == 1 && isError(elements[0]) {
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

// ____________
//
// eval Expression
// ____________

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, exp := range exps {
		evaled := Eval(exp, env)
		if isError(evaled) {
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
		return newError("unknown operator: %s%s", loc, operator, right.Type())
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
		return newError("unknown operator: -%s", loc, right.Type())
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
		return newError("unknown operator: %s %s %s", loc, left.Type(), operator, right.Type())
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
		return newError("unknown operator: %s %s %s", loc, left.Type(), operator, right.Type())
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
		return newError("type mismatch: %s %s %s", loc, left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", loc, left.Type(), operator, right.Type())
	}
}

func evalArrayIndexExpression(arr object.Object, idx object.Object, loc *token.TokenLocation) object.Object {
	ao, ok := arr.(*object.Array)
	if !ok {
		return newError("type mismatch for array index operation. got=%s", loc, arr.Type())
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
		return newError("type mismatch for hash index operation. got=%s", loc, hash.Type())

	}

	hashKey, ok := idx.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", loc, idx.Type())
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
		return newError("index operator not supported: %s", loc, left.Type())
	}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)

	if isTruthy(condition) {
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
	return newError("identifier not found: "+node.Value, &node.Token.Location)
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)
	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", &node.Token.Location, key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
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

// isTruthy returns false only if object is NULL or FALSE, everything else is true
func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case FALSE:
		return false
	case TRUE:
		return true
	default:
		return true
	}
}

func newError(format string, location *token.TokenLocation, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...), Location: location}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func applyFunction(fn object.Object, loc *token.TokenLocation, args []object.Object) object.Object {
	switch fun := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fun, args)
		evaled := Eval(fun.Body, extendedEnv)
		return unwrapReturnValue(evaled)

	case *object.Builtin:
		return fun.Fn(args...)

	default:
		return newError("not a function: %s", loc, fn.Type())
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
