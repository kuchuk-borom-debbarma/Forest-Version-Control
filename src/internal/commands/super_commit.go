package commands

import (
	v1 "MultiRepoVC/src/internal/core/version_control/v1"
)

type SuperCommitCommand struct {
	BaseCommand
}

func (c *SuperCommitCommand) Name() string { return "super-commit" }

func (c *SuperCommitCommand) Description() string {
	return "Creates a hierarchical super commit for this repo and its children."
}

func (c *SuperCommitCommand) RequiredArgs() []string { return []string{"message"} }
func (c *SuperCommitCommand) OptionalArgs() []string { return []string{"author"} }

func (c *SuperCommitCommand) ExecuteCommand(p map[string][]string) error {
	message := p["message"][0]

	author := "unknown"
	if a, ok := p["author"]; ok && len(a) > 0 {
		author = a[0]
	}

	vc := v1.New()
	return vc.SuperCommit(message, author)
}

func init() {
	Global.Register(&SuperCommitCommand{})
}
