[![CI](https://github.com/drdreo/donkey-script/actions/workflows/go.yml/badge.svg)](https://github.com/drdreo/donkey-script/actions/workflows/go.yml)

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
[] add http GET request, make it blocking, then non-blocking with go routines
[] add import files support
