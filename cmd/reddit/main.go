// Command reddit is a CLI for reading public Reddit data.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/fang"
	"github.com/tamnd/reddit-cli/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := cli.NewRootCmd()
	err := fang.Execute(ctx, root,
		fang.WithVersion(cli.Version),
		fang.WithNotifySignal(os.Interrupt, syscall.SIGTERM),
	)
	if err == nil {
		return
	}

	var ee *cli.ExitError
	if errors.As(err, &ee) {
		if ee.Err != nil {
			fmt.Fprintln(os.Stderr, "reddit:", ee.Err)
		}
		os.Exit(ee.Code)
	}
	fmt.Fprintln(os.Stderr, "reddit:", err)
	os.Exit(1)
}
