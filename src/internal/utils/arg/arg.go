package arg

import "strings"

// ParseArgs converts CLI args into a key → []values map.
//
// Supports:
//
//	--key value1 value2 value3
//	--flag
//	positional values
//	--key=value
//
// All non-flag values following a flag are grouped under it
// until the next --flag is found.
func ParseArgs(args []string) map[string][]string {
	result := make(map[string][]string)

	currentKey := "positional"

	for i := 0; i < len(args); i++ {
		token := args[i]

		// Case: --key=value
		if strings.HasPrefix(token, "--") && strings.Contains(token, "=") {
			parts := strings.SplitN(token[2:], "=", 2)
			key := parts[0]
			value := parts[1]

			result[key] = append(result[key], value)
			currentKey = key
			continue
		}

		// Case: --flag or --key
		if strings.HasPrefix(token, "--") {
			key := token[2:]

			// Next item is a value unless it is another flag
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				// Assign upcoming values to this key
				currentKey = key
				continue
			}

			// No value → treat as boolean flag
			result[key] = append(result[key], "true")
			currentKey = key
			continue
		}

		// Case: a non-flag value → belongs to currentKey
		result[currentKey] = append(result[currentKey], token)
	}

	return result
}
