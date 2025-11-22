package commands

import (
	v1 "MultiRepoVC/src/internal/core/version_control/v1"
)

type InitCommand struct {
	BaseCommand
}

func (c *InitCommand) Name() string { return "init" }
func (c *InitCommand) Description() string {
	return "Initializes a new MRVC repository."
}
func (c *InitCommand) RequiredArgs() []string { return []string{"name", "author"} }
func (c *InitCommand) OptionalArgs() []string { return []string{} }

func (c *InitCommand) ExecuteCommand(p map[string][]string) error {
	name := p["name"][0]
	author := p["author"][0]

	vc := v1.New()
	return vc.Init(name, author)
}

func init() {
	Global.Register(&InitCommand{})
}
