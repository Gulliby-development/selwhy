package cli

import (
	"errors"

	"github.com/Gulliby-development/selwhy/internal/base"
	"github.com/Gulliby-development/selwhy/internal/selector"
)

// runFmt prints the canonical form of a selector list.
func runFmt(args []string, io base.IO) (int, error) {
	parsed, err := base.Parse(args, nil)
	if err != nil {
		return 0, err
	}
	if len(parsed.Args) == 0 {
		return 0, errors.New("missing selector")
	}
	if len(parsed.Args) > 1 {
		return 0, errors.New("too many arguments")
	}

	text, err := base.TextArg(io, parsed.Args[0])
	if err != nil {
		return 0, err
	}
	list, err := selector.Parse(text)
	if err != nil {
		return 0, err
	}

	formatted := base.Show(selector.ListText(list))
	formatted = formatted[1 : len(formatted)-1]
	io.Out(formatted + "\n")
	return 0, nil
}
