package main

import (
	"context"
	"fmt"
	"os"

	"github.com/shuldan/cli"

	"github.com/shuldan/skeleton/internal/bootstrap"
)

func main() {
	if err := bootstrap.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.GetExitCode(err))
	}
}
