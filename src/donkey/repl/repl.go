package repl

import (
	"bufio"
	"donkey/compiler"
	"donkey/compiler/vm"
	"donkey/constants"
	"donkey/lexer"
	"donkey/parser"
	"fmt"
	"io"
)

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	//	env := object.NewEnvironment()
	//	macroEnv := object.NewEnvironment()

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

		// COMPILER section
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			printCompilerErrors(out, []string{err.Error()})
			continue
		}

		machine := vm.New(comp.Bytecode())
		err = machine.Run()

		if err != nil {
			printCompilerErrors(out, []string{fmt.Sprintf("Bytecode execution failed:\n%s\n", err)})
			continue
		}

		stackTop := machine.StackTop()
		io.WriteString(out, stackTop.Inspect())
		io.WriteString(out, "\n")

		// ----
		// EVAL section

		// support macros in REPL
		//		evaluator.DefineMacros(program, macroEnv)
		//		expanded := evaluator.ExpandMacros(program, macroEnv)
		//
		//		evaled := evaluator.Eval(expanded, env)
		//		if evaled != nil {
		//			io.WriteString(out, evaled.Inspect())
		//			io.WriteString(out, "\n")
		//		}
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, constants.ParserErrorPrompt)
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}

func printCompilerErrors(out io.Writer, errors []string) {
	io.WriteString(out, constants.CompilerErrorPrompt)
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
