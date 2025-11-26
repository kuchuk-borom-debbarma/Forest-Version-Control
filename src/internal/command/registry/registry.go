package registry

import (
	"fmt"
	"main/internal/command"
)

// commandRegistry holds all registered commands. Commands should register themselves here.
var commandRegistry = make(map[string]command.Command)
var commandExecutor = CommandExecutor{}

func Execute(name string, args []string) error {
	command, exists := commandRegistry[name]
	if !exists {
		fmt.Println("Invalid command!")
	}
	var argsMap map[string]any //TODO
	return commandExecutor.executeCommand(command, argsMap)
}

func Register(command command.Command) {
	name := command.Name()
	commandRegistry[name] = command
}

type CommandExecutor struct{}

func (bc *CommandExecutor) executeCommand(c command.Command, args map[string]any) error {
	err := c.IsArgsValid(args)
	if err != nil {
		return err
	}
	return c.Execute(args)
}
