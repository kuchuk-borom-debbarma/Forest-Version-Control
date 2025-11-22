package commands

type Command interface {
	Name() string
	Description() string
	RequiredArgs() []string
	OptionalArgs() []string

	// ExecuteCommand parsed: key → []values
	ExecuteCommand(parsed map[string][]string) error
}
