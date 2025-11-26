package command

type Command interface {
	// Name returns the name of the command
	Name() string
	// Description returns a brief description of the command
	Description() string
	// IsArgsValid validates the provided arguments for the command.
	// This method avoids relying solely on predefined optional/required arguments,
	// since some arguments are only valid in specific combinations rather than individually.
	// It allows each command to define complex validation rules tailored to its logic.
	IsArgsValid(args map[string]any) error
	// Execute runs the command with the provided arguments.
	Execute(args map[string]any) error
}
