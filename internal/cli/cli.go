package cli

import "github.com/Gulliby-development/selwhy/internal/base"

// Usage lists the supported commands in dispatch order.
const Usage = "selwhy tree DOC\n" +
	"selwhy select SELECTOR DOC\n" +
	"selwhy explain SELECTOR DOC N\n" +
	"selwhy specificity SELECTOR\n" +
	"selwhy fmt SELECTOR\n" +
	"selwhy nth [--count N] FORMULA\n" +
	"selwhy cascade SHEET DOC\n"

var commands = map[string]func([]string, base.IO) (int, error){
	"tree":        runTree,
	"select":      runSelect,
	"explain":     runExplain,
	"specificity": runSpecificity,
	"fmt":         runFmt,
	"nth":         runNth,
	"cascade":     runCascade,
}

// Run dispatches argv to a command, printing a single prefixed diagnostic on error.
func Run(argv []string, io base.IO) int {
	if len(argv) == 0 {
		io.Err("selwhy: missing command (tree, select, explain, specificity, fmt, nth or cascade)\n")
		return 2
	}

	if len(argv) == 1 && argv[0] == "--help" {
		io.Out(Usage)
		return 0
	}

	run, ok := commands[argv[0]]
	if !ok {
		io.Err("selwhy: unknown command " + base.Show(argv[0]) + "\n")
		return 2
	}

	code, err := run(argv[1:], io)
	if err != nil {
		io.Err("selwhy: " + err.Error() + "\n")
		return 2
	}
	if code != 0 && code != 1 {
		return 2
	}
	return code
}
