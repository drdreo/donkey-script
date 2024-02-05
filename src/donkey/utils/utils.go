package utils

import (
	"donkey/ast"
	"donkey/lexer"
	"donkey/parser"
)

func ParseProgram(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}
