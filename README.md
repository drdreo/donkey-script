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
[ ] make http fetch non-blocking with go routines
[ ] add import files support
