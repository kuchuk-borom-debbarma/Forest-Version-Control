package commands

import (
	"errors"

	v1 "MultiRepoVC/src/internal/core/version_control/v1"
	"MultiRepoVC/src/internal/utils/arg"
)

type InitCommand struct{}

func (c *InitCommand) Name() string {
	return "init"
}

func (c *InitCommand) Execute(args []string) error {
	parsedArgs := arg.ParseArgs(args)

	// 1. Validate required arguments
	name, ok := parsedArgs["name"]
	if !ok || name == "" {
		return errors.New("missing required argument: --name")
	}

	author, ok := parsedArgs["author"]
	if !ok || author == "" {
		return errors.New("missing required argument: --author")
	}

	// 2. Execute VC Init
	vc := v1.New()
	return vc.Init(name, author)
}

// auto-register this command
func init() {
	Register(&InitCommand{})
}
