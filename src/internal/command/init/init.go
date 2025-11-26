package init

import (
	"fmt"
	"main/internal/command/registry"
	"main/internal/versioncontrol"
	"strings"
)

type Command struct{}

func (i *Command) Name() string {
	return "init"
}
func (i *Command) Description() string {
	return "Initialize a directory as Forest Version Control."
}
func (i *Command) IsArgsValid(args map[string]any) error {
	var misses []string

	if _, exists := args["name"]; !exists {
		misses = append(misses, "--name param missing")
	}
	if _, exists := args["author"]; !exists {
		misses = append(misses, "--author param is missing")
	}

	if len(misses) > 0 {
		return fmt.Errorf(strings.Join(misses, "\n"))
	}

	return nil
}
func (i *Command) Execute(args map[string]any) error {
	repoName, exists := args["name"].(string)
	author, exists := args["author"].(string)
	if !exists || author == "" {
		author = "unknown"
	}

	return versioncontrol.Init(repoName, author)
}

// Invoked when package is initialized. Automatically registering
func init() {
	registry.Register(&Command{})
}
