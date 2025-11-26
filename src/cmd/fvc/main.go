package main

import (
	"fmt"
	"main/internal/command/registry"
	"os"
)

func main() {
	fmt.Print("Forest Version Control")
	args := os.Args[1:] // skip program name
	if len(args) == 0 {
		fmt.Println("No command provided")
		return
	}

	name := args[0]     // the command (e.g. "init")
	cmdArgs := args[1:] // arguments after command

	err := registry.Execute(name, cmdArgs)
	if err != nil {
		return
	}
}
