package main

import (
	"MultiRepoVC/src/internal/commands"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No command provided.")
		fmt.Println("Use 'mrvc help' to see available commands.")
		return
	}

	cmdName := os.Args[1]

	if cmdName == "help" {
		commands.Global.List()
		return
	}

	cmd, err := commands.Global.Get(cmdName)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println()
		fmt.Println("Use 'mrvc help' to see available commands.")
		return
	}

	base := commands.BaseCommand{}
	if err := base.Run(cmd, os.Args[2:]); err != nil {
		fmt.Println("Error:", err)
	}
}
