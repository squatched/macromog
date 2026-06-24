package main

import "fmt"

// parseConfigFlags extracts boolean and string flags from args, returning the
// remaining positional arguments in their original order. This allows flags to
// appear before or after positional arguments (e.g. add-install steam /path --set-default).
func parseConfigFlags(args []string, stringFlags map[string]*string) (setDefault bool, positional []string, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--set-default":
			setDefault = true
		case "--install":
			if stringFlags == nil || stringFlags["install"] == nil {
				return false, nil, fmt.Errorf("macromog config: unexpected flag --install")
			}
			if i+1 >= len(args) {
				return false, nil, fmt.Errorf("macromog config: --install requires a value")
			}
			i++
			*stringFlags["install"] = args[i]
		default:
			positional = append(positional, arg)
		}
	}
	return setDefault, positional, nil
}
