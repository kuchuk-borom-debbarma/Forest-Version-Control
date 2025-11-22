package commands

import (
	"fmt"
	"strings"
)

var Global = NewRegistry()

type Registry struct {
	commands map[string]Command
}

func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

func (r *Registry) Get(name string) (Command, error) {
	cmd, ok := r.commands[name]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s", name)
	}
	return cmd, nil
}

func (r *Registry) List() {
	fmt.Println("\nAvailable commands:")
	for name, cmd := range r.commands {
		fmt.Printf("  %s  -  %s\n", name, cmd.Description())
		fmt.Printf("      Required: --%s\n", strings.Join(cmd.RequiredArgs(), " --"))
		if len(cmd.OptionalArgs()) > 0 {
			fmt.Printf("      Optional: --%s\n", strings.Join(cmd.OptionalArgs(), " --"))
		}
		fmt.Println()
	}
}
