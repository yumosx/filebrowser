package main

import (
	"log/slog"
	"os"

	"github.com/filebrowser/filebrowser/v2/cmd"
	"github.com/filebrowser/filebrowser/v2/errors"
)

func main() {
	slog.Info("start the server")
	if err := cmd.Execute(); err != nil {
		os.Exit(errors.GetExitCode(err))
	}
}
