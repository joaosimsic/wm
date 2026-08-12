package main

import (
	"github.com/joaosimsic/wm/internal/x11"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	logger.Info("Starting WM")

	conn, err := x11.Connect()
	if err != nil {
		logger.Fatal("failed to connect to x11", zap.Error(err))
	}
	defer conn.Close()

	logger.Info("connected to X11")
}
