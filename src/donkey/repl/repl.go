package repl

import (
	"bufio"
	"donkey/constants"
	"donkey/evaluator"
	"donkey/lexer"
	"donkey/object"
	"donkey/parser"
	"fmt"
	"io"
)

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Fprintf(out, constants.ReplPrompt)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaled := evaluator.Eval(program, env)
		if evaled != nil {
			io.WriteString(out, evaled.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, constants.ParserErrorPrompt)
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
