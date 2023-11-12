package repl

import (
	"bufio"
	"donkey/constants"
	"donkey/lexer"
	"donkey/parser"
	"fmt"
	"io"
)

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

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

		io.WriteString(out, program.String())
		io.WriteString(out, "\n")

	}
}

const DONKEY_FACE = `
_\
/b
/####J
|\ || Yikes
`

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, DONKEY_FACE)
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
