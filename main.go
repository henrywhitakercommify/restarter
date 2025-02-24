package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/henrywhitakercommify/restarter/cmd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx, cancel = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGKILL)
	defer cancel()

	cmd := cmd.NewRoot()
	cmd.SetContext(ctx)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
