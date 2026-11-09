package main

import (
	"honi/internal/cmd"
	"honi/internal/log"
	"log/slog"
	"os"
)

func main() {
	if err := log.ConfigLog(); err != nil {
		slog.Error("ConfigLog failed", "err", err)
		os.Exit(1)
	}
	cmd.Execute()
}
