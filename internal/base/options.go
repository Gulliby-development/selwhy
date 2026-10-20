package base

import (
	"errors"
	"strings"
)

// Spec describes an option accepted by Parse.
type Spec struct {
	Name string
}

// Parsed contains option values and positional arguments in encounter order.
type Parsed struct {
	Values map[string][]string
	Args   []string
}

// Parse reads options and ordinary arguments from left to right.
func Parse(argv []string, specs []Spec) (Parsed, error) {
	parsed := Parsed{Values: make(map[string][]string)}
	known := make(map[string]struct{}, len(specs))
	for _, spec := range specs {
		known[spec.Name] = struct{}{}
	}

	optionsEnded := false
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if optionsEnded {
			parsed.Args = append(parsed.Args, arg)
			continue
		}
		if arg == "--" {
			optionsEnded = true
			continue
		}
		if !strings.HasPrefix(arg, "--") {
			parsed.Args = append(parsed.Args, arg)
			continue
		}

		name := arg
		value := ""
		hasEquals := false
		if equals := strings.IndexByte(arg, '='); equals >= 0 {
			name = arg[:equals]
			value = arg[equals+1:]
			hasEquals = true
		}
		if _, ok := known[name]; !ok {
			return parsed, errors.New("unknown option " + Show(arg))
		}
		if !hasEquals {
			if i+1 >= len(argv) {
				return parsed, errors.New(name + " needs a value")
			}
			i++
			value = argv[i]
		}
		parsed.Values[name] = append(parsed.Values[name], value)
	}
	return parsed, nil
}
