package log

import (
	"honi/config"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func ConfigLog() error {
	conf := config.GetConfig()
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("ConfigLog UserHomeDir failed", "err", err)
		return err
	}
	logPath := filepath.Join(home, ".local", "state", "honi", "honi.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		slog.Error("ConfigLog MkdirAll failed", "err", err)
		return err
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("ConfigLog OpenFile failed", "err", err)
		return err
	}

	level := slog.LevelInfo
	if conf.Debug {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(io.MultiWriter(os.Stdout, file), &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
	return nil
}
