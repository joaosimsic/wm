package main

import (
	"github.com/joaosimsic/wm/internal/config"
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

	cfg, err := config.Load()
	if err != nil {
		logger.Warn("config issues, using defaults where possible", zap.Error(err))
	}

	conn, err := x11.Connect()
	if err != nil {
		logger.Fatal("failed to connect to x11", zap.Error(err))
	}
	defer conn.Close()

	logger.Info("connected to X11")

	colors, err := cfg.Colors.RGB()
	if err != nil {
		logger.Fatal("failed to parse colors", zap.Error(err))
	}

	pixel, err := conn.AllocColor(
		uint16(colors.BorderActive.R),
		uint16(colors.BorderActive.G),
		uint16(colors.BorderActive.B),
	)
	if err != nil {
		logger.Fatal("failed to allocate color", zap.Error(err))
	}

	logger.Info("config loaded",
		zap.String("terminal", cfg.Terminal),
		zap.String("font", cfg.Font),
		zap.Uint32("border_active_pixel", pixel),
	)
}
