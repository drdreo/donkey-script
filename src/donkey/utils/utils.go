package utils

import (
	"fmt"

	"donkey/ast"
	"donkey/lexer"
	"donkey/parser"
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
