package compiler

import (
	"donkey/ast"
	"donkey/compiler/code"
	"donkey/compiler/symbol"
	"donkey/object"
	"fmt"
	"sort"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

type CompilationScope struct {
	instructions    code.Instructions
	lastInstruction EmittedInstruction
	prevInstruction EmittedInstruction
}

type Compiler struct {
	constants   []object.Object
	symbolTable *symbol.SymbolTable
	scopes      []CompilationScope
	scopeIdx    int
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:    code.Instructions{},
		lastInstruction: EmittedInstruction{},
		prevInstruction: EmittedInstruction{},
	}
	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbol.NewSymbolTable(),
		scopes:      []CompilationScope{mainScope},
		scopeIdx:    0,
	}
}

func NewWithState(s *symbol.SymbolTable, constants []object.Object) *Compiler {
	// re-allocation is fine, i guess?
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)

	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		sym := c.symbolTable.Define(node.Name.Value)
		c.emit(code.OpSetGlobal, sym.Index)

	case *ast.Identifier:
		sym, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.emit(code.OpGetGlobal, sym.Index)

	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.InfixExpression:
		var leftNode = node.Left
		var rightNode = node.Right

		if reorderExpression(node.Operator) {
			leftNode = node.Right
			rightNode = node.Left
		}

		err := c.Compile(leftNode)
		if err != nil {
			return err
		}

		err = c.Compile(rightNode)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSubtract)
		case "*":
			c.emit(code.OpMultiply)
		case "/":
			c.emit(code.OpDivide)

		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		case ">", "<":
			c.emit(code.OpGreaterThan)
		case ">=", "<=":
			c.emit(code.OpGreaterThanOrEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit an `OpJumpNotTruthy` with a garbage value, back-patched
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.isLastInstruction(code.OpPop) {
			c.removeLastInstruction()
		}

		// Emit an `OpJump` with a garbage vlaue, back-patched
		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		// no else-block, emit a NULL instead
		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err = c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.isLastInstruction(code.OpPop) {
				c.removeLastInstruction()
			}
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.BooleanLiteral:
		op := code.OpFalse
		if node.Value {
			op = code.OpTrue
		}
		c.emit(op)

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))

	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))

	case *ast.HashLiteral:
		var keys []ast.Expression
		for k := range node.Pairs {
			keys = append(keys, k)
		}
		// sort hash keys for predictability in tests
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}
			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}
		c.emit(code.OpHash, len(node.Pairs)*2)

	case *ast.FunctionLiteral:
		c.enterScope()

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		// to support implicit returns
		// e.g. fn(){5}
		if c.isLastInstruction(code.OpPop) {
			c.replaceLastInstruction(code.OpReturnValue)
		}
		// empty functions, that dont return anything, implicitly nor explicitly
		// e.g. fn(){}
		if !c.isLastInstruction(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		instructions := c.leaveScope()

		compiledFunc := &object.CompiledFunction{Instructions: instructions}
		c.emit(code.OpConstant, c.addConstant(compiledFunc))

	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)

	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		c.emit(code.OpCall)

	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)

	}
	return nil
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:    code.Instructions{},
		lastInstruction: EmittedInstruction{},
		prevInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIdx++
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIdx--

	return instructions
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIdx].instructions
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)

	c.scopes[c.scopeIdx].instructions = updatedInstructions

	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	c.scopes[c.scopeIdx].prevInstruction = c.scopes[c.scopeIdx].lastInstruction
	c.scopes[c.scopeIdx].lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
}

func (c *Compiler) isLastInstruction(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIdx].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastInstruction() {
	last := c.scopes[c.scopeIdx].lastInstruction
	prev := c.scopes[c.scopeIdx].prevInstruction

	curInstructions := c.currentInstructions()
	newInstructions := curInstructions[:last.Position]

	c.scopes[c.scopeIdx].instructions = newInstructions
	c.scopes[c.scopeIdx].lastInstruction = prev
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	curInstructions := c.currentInstructions()
	// this assumes non-variable, same length instructions
	for i := 0; i < len(newInstruction); i++ {
		curInstructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) replaceLastInstruction(op code.Opcode) {
	opcode := code.Make(op)
	lastPos := c.scopes[c.scopeIdx].lastInstruction.Position
	c.replaceInstruction(lastPos, opcode)
	c.scopes[c.scopeIdx].lastInstruction.Opcode = op
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	curInstructions := c.currentInstructions()

	op := code.Opcode(curInstructions[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func reorderExpression(operator string) bool {
	return operator == "<" || operator == "<="
}
