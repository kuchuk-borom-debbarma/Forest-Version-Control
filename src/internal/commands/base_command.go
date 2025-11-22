package commands

import (
	"MultiRepoVC/src/internal/utils/arg"
	"fmt"
)

// BaseCommand Template method providing common behavior
type BaseCommand struct{}

func (b *BaseCommand) Run(cmd Command, args []string) error {
	parsed := arg.ParseArgs(args)

	// Validate required args
	for _, req := range cmd.RequiredArgs() {
		values, exists := parsed[req]
		if !exists || len(values) == 0 {
			return fmt.Errorf("missing required argument: --%s", req)
		}
	}

	return cmd.ExecuteCommand(parsed)
}
