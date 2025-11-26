package commit

import (
	"fmt"
	"main/internal/command/registry"
	"main/internal/versioncontrol"
	"strings"
)

type Command struct{}

func (i *Command) Name() string {
	return "commit"
}
func (i *Command) Description() string {
	return "Initialize a directory as Forest Version Control."
}
func (i *Command) IsArgsValid(args map[string]any) error {
	var misses []string

	if _, exists := args["message"]; !exists {
		misses = append(misses, "--message param missing")
	}

	if len(misses) > 0 {
		return fmt.Errorf(strings.Join(misses, "\n"))
	}

	return nil
}
func (i *Command) Execute(args map[string]any) error {
	message, exists := args["message"].(string)
	author, exists := args["author"].(string)
	if !exists || author == "" {
		author = "unknown"
	}
	files := args["files"].([]string)

	return versioncontrol.Commit(message, author, files)
}

// Invoked when package is initialized. Automatically registering
func init() {
	registry.Register(&Command{})
}
