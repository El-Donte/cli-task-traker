package main

import (
	"cli-task-traker/internal/cli"
	"cli-task-traker/internal/service"
	"cli-task-traker/internal/storage"
	"fmt"
	"os"
)

const storagePath = "tasks.json"

func main() {
	if err := storage.EnsureDir(storagePath); err != nil {
		os.Exit(1)
	}

	store, err := storage.OpenFileStorage(storagePath)

	if err != nil {
		os.Exit(1)
	}

	svc := service.NewService(store)
	app := cli.NewCLI(svc, os.Stdin, os.Stdout)
	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
