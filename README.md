[![CI](https://github.com/drdreo/donkey-script/actions/workflows/go.yml/badge.svg)](https://github.com/drdreo/donkey-script/actions/workflows/go.yml)

# REPL
Run `go run .` in `/src/donkey` to execute the REPL.

## Testing
runnings test coverage
`go test -coverpkg=./... ./...`


## Error Handling

[x] added line and column numbers
[ ] show code where error occured

## Strings

[x] string concatenation:   `"he" + "yo" = "heyo"`
[x] string substraction:    `"hey ho there" - "ho" = "hey  there"`
[x] string equal:           `"hey" == "hey" = true`

## Numbers

[ ] fix edge case arithmetic : `1 / 0`


## Builtins
[x] add blocking http GET request
[x] make http fetch non-blocking with go routines
[ ] add import files support


## Background

### Routine
We interpret Donkey source code in a series of steps.
First comes the lexer which turns source code it into tokens.
Then comes the parser and turns the tokens into an AST.
Afterwards, macros are processed and expand the AST.
Finally, `Eval` takes this AST and evaluates its nodes recursively, statement by statement, expression by expression.

Lexing, parsing (macro expansion), import resolution, and evaluation -- strings to tokens, tokens to AST, macro expansion, AST to output.
