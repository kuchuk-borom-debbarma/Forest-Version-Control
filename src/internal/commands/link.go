package commands

import (
	v1 "MultiRepoVC/src/internal/core/version_control/v1"
)

type LinkCommand struct{}

func (c *LinkCommand) ExecuteCommand(parsed map[string][]string) error {
	toLink := parsed["path"][0]
	vcs := v1.New()
	err := vcs.Link(toLink)
	if err != nil {
		return err
	}
	return nil
}

func (c *LinkCommand) Name() string {
	return "link"
}

func (c *LinkCommand) Description() string {
	return "Link a child repo to the current repo"
}

func (c *LinkCommand) RequiredArgs() []string { return []string{"path"} }
func (c *LinkCommand) OptionalArgs() []string { return []string{} }

func init() {
	Global.Register(&LinkCommand{})
}
