package main

import (
	"io"
	"os"

	"github.com/Gulliby-development/selwhy/internal/base"
	"github.com/Gulliby-development/selwhy/internal/cli"
)

type standardIO struct{}

func (standardIO) ReadStdin() []byte {
	data, _ := io.ReadAll(os.Stdin)
	return data
}

func (standardIO) Out(text string) {
	_, _ = os.Stdout.Write([]byte(text))
}

func (standardIO) Err(text string) {
	_, _ = os.Stderr.Write([]byte(text))
}

var _ base.IO = standardIO{}

func main() {
	os.Exit(cli.Run(os.Args[1:], standardIO{}))
}
