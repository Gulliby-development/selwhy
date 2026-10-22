package cli

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Gulliby-development/selwhy/internal/base"
	"github.com/Gulliby-development/selwhy/internal/markup"
)

func runTree(args []string, io base.IO) (int, error) {
	parsed, err := base.Parse(args, []base.Spec{{Name: "--depth"}})
	if err != nil {
		return 0, err
	}

	limit := 12
	if values := parsed.Values["--depth"]; len(values) > 0 {
		limit, err = parseTreeDepth(values[len(values)-1])
		if err != nil {
			return 0, err
		}
	}

	if len(parsed.Args) == 0 {
		return 0, errors.New("missing document")
	}
	if len(parsed.Args) > 1 {
		return 0, errors.New("too many arguments")
	}

	text, err := base.TextArg(io, parsed.Args[0])
	if err != nil {
		return 0, err
	}
	doc, err := markup.Parse(text)
	if err != nil {
		return 0, err
	}

	rows := make([][]string, 0, len(doc.Elements))
	for _, element := range doc.Elements {
		if markup.Depth(element) > limit {
			continue
		}
		rows = append(rows, []string{
			strconv.Itoa(element.Num),
			markup.Path(element),
			base.Cell(markup.TagText(element)),
		})
	}

	lines := base.Table([]string{"n", "path", "tag"}, rows)
	lines = append(lines, base.Line("elements", strconv.Itoa(len(rows))))
	io.Out(strings.Join(lines, "\n") + "\n")
	return 0, nil
}

// parseTreeDepth validates the bounded ASCII depth syntax before conversion.
func parseTreeDepth(text string) (int, error) {
	valid := len(text) >= 1 && len(text) <= 2 && text[0] != '0'
	for i := 0; valid && i < len(text); i++ {
		valid = text[i] >= '0' && text[i] <= '9'
	}
	if valid {
		value, err := strconv.Atoi(text)
		if err == nil && value >= 1 && value <= 12 {
			return value, nil
		}
	}
	return 0, errors.New("invalid --depth " + base.Show(text) + " (expected a whole number from 1 to 12)")
}
