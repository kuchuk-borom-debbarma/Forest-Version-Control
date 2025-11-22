package commands

import (
	v1 "MultiRepoVC/src/internal/core/version_control/v1"
	"errors"
)

type CommitCommand struct {
	BaseCommand
}

func (c *CommitCommand) Name() string { return "commit" }
func (c *CommitCommand) Description() string {
	return "Creates a new commit with a message and files."
}

func (c *CommitCommand) RequiredArgs() []string { return []string{"message"} }
func (c *CommitCommand) OptionalArgs() []string { return []string{"author", "files"} }

func (c *CommitCommand) ExecuteCommand(p map[string][]string) error {
	message := p["message"][0]

	author := "unknown"
	if a, ok := p["author"]; ok && len(a) > 0 {
		author = a[0]
	}

	// PRIMARY way: --files file1 file2 file3
	files := p["files"]

	// FALLBACK: positional arguments
	if len(files) == 0 {
		files = p["positional"]
	}

	if len(files) == 0 {
		return errors.New("no files specified")
	}

	vc := v1.New()
	return vc.Commit(message, author, files)
}

func init() {
	Global.Register(&CommitCommand{})
}
