package main

import (
	"donkey/constants"
	"fmt"
	"os/user"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is the %s programming language!\n",
		user.Username, constants.DonkeyLangName)
	fmt.Printf("Feel free to type in commands\n")
}
