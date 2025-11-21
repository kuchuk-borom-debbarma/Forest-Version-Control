package commands

import "fmt"

var registry = map[string]Command{}

func Register(cmd Command) {
	name := cmd.Name()

	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("command %s already registered", name))
	}

	registry[name] = cmd
}

func Get(name string) (Command, bool) {
	cmd, ok := registry[name]
	return cmd, ok
}

func All() map[string]Command {
	return registry
}
