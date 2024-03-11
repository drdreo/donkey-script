package repl

import (
	"bufio"
	"donkey/compiler"
	"donkey/compiler/symbol"
	"donkey/compiler/vm"
	"donkey/constant"
	"donkey/lexer"
	"donkey/object"
	"donkey/parser"
	"fmt"
	"io"
)

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	// 	EVAL envs
	//	env := object.NewEnvironment()
	//	macroEnv := object.NewEnvironment()

	// COMPILER envs
	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalSize)
	symbolTable := symbol.NewSymbolTable()

	for {
		fmt.Fprintf(out, constant.ReplPrompt)
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
		comp := compiler.NewWithState(symbolTable, constants)
		err := comp.Compile(program)
		if err != nil {
			printCompilerErrors(out, []string{err.Error()})
			continue
		}

		code := comp.Bytecode()
		//	huh?	constants = code.Constants
		machine := vm.NewWithGlobalState(code, globals)

		err = machine.Run()
		if err != nil {
			printCompilerErrors(out, []string{fmt.Sprintf("Bytecode execution failed:\n%s\n", err)})
			continue
		}

		stackTop := machine.LastPoppedStackElem()
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
	io.WriteString(out, constant.ParserErrorPrompt)
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}

func printCompilerErrors(out io.Writer, errors []string) {
	io.WriteString(out, constant.CompilerErrorPrompt)
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
