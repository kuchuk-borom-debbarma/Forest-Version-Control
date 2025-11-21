package arg

// ParseArgs converts CLI args into a key-value map.
// Supports:
//
// mrvc init --name=test --author="kuku"
// mrvc commit --msg "hello" file1 file2
func ParseArgs(args []string) map[string]string {
	parsed := make(map[string]string)

	positionalIndex := 0

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// --key=value
		if len(arg) > 2 && arg[:2] == "--" && contains(arg, "=") {
			parts := split(arg[2:], "=")
			parsed[parts[0]] = parts[1]
			continue
		}

		// --flag or --key value
		if len(arg) > 2 && arg[:2] == "--" {
			key := arg[2:]

			// Check next token
			if i+1 < len(args) && !isFlag(args[i+1]) {
				parsed[key] = args[i+1]
				i++
			} else {
				// flag without value
				parsed[key] = "true"
			}
			continue
		}

		// positional argument
		parsed[string(rune(positionalIndex))] = arg
		positionalIndex++
	}

	return parsed
}

// helpers -----------------------------

func contains(s, substr string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == substr[0] {
			return true
		}
	}
	return false
}

func split(s, sep string) []string {
	out := make([]string, 2)
	idx := -1
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			idx = i
			break
		}
	}
	if idx == -1 {
		out[0] = s
		out[1] = ""
		return out
	}
	out[0] = s[:idx]
	out[1] = s[idx+1:]
	return out
}

func isFlag(s string) bool {
	return len(s) > 2 && s[:2] == "--"
}
