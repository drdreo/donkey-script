package main

import (
	"donkey/constant"
	"donkey/repl"
	"fmt"
	"os"
	"os/user"
)

func main() {
	userr, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is the %s programming language!\n",
		userr.Username, constant.LangName)
	fmt.Printf("Feel free to type in commands\n")

	repl.Start(os.Stdin, os.Stdout)

	// clear console colors on close
	defer func() { fmt.Println("\u001b[39m") }()
}
