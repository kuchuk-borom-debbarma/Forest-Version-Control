package main

import (
	"fmt"
	"os"

	"MultiRepoVC/src/internal/commands"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mrvc <command> [--flags]")
		return
	}

	cmdName := os.Args[1]

	cmd, ok := commands.Get(cmdName)
	if !ok {
		fmt.Println("Unknown command:", cmdName)
		return
	}

	if err := cmd.Execute(os.Args[2:]); err != nil {
		fmt.Println("Error:", err)
	}
}
