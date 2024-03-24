[![CI](https://github.com/drdreo/donkey-script/actions/workflows/go.yml/badge.svg)](https://github.com/drdreo/donkey-script/actions/workflows/go.yml)

# REPL
Run `go run .` in `/src/donkey` to execute the REPL.

## Testing
runnings test coverage
`go test -coverpkg=./... ./...`


## Features

 - `< > == != <= >=` comparison operators are supported

## If

### Assignments
`let x = if(true){69}; // x == 69`

### Nested Expressions
`if ((if (false) { 13 })) {
    69;
} else {
    420;
}`

## Booleans

[x][x] boolean expressions: `(1 > 2) == false`

## Strings

[x][x] string concatenation:   `"he" + "yo" = "heyo"`
[x][x] string substraction:    `"hey ho there" - "ho" = "hey  there"`
[x][x] string comparison:      `"hey" == "hey" = true`

## Numbers

[ ][ ] fix edge case arithmetic : `1 / 0`

## Arrays

[x][x] index access:           `[420][2-1] = 420`
[x][x] negative index access:  `[1,2,3][-1] = 3`
[ ][ ] array concatenation:    `[1,2,3] + [420] = [1,2,3,420]`

## Hashes

[x][x] index access:           `{1:420}[2-1] = 420`
[ ][ ] hash concatenation:     `{1:2} + {2: 3} = {1:2, 2:3}`

## Builtins
[x][ ] add blocking http GET request
[x][ ] make http fetch non-blocking with go routines
[ ][ ] add import files support

## Error Handling

[x][ ] added line and column numbers
[ ][ ] show code where error occured


## Background

### Interpreter
We interpret Donkey source code in a series of steps.
First comes the lexer which turns source code it into tokens.
Then comes the parser and turns the tokens into an AST.
Afterwards, macros are processed and expand the AST.
Finally, `Eval` takes this AST and evaluates its nodes recursively, statement by statement, expression by expression.

#### Flow
Lexing, parsing (macro expansion) and evaluation

String --> Tokens --> AST --> Macro Expansion  --> Objects (output)

### Compiler
As the virtual machine, a stack machine was chosen over registers.
Why? Easier to grasp and build.

#### Flow
Lexer --> Parser --> Compiler --> Virtual Machine

String --> Tokens --> AST --> Bytecode --> Objects
