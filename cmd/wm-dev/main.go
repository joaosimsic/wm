package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
)

const (
	display = ":1"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	logger.Info("starting Xephyr", zap.String("display", display))

	xephyr := exec.CommandContext(
		ctx,
		"Xephyr",
		"-br",
		"-ac",
		"-noreset",
		"-screen",
		"1920x1080",
		display,
	)

	xephyrStdout, err := xephyr.StdoutPipe()
	if err != nil {
		logger.Fatal("failed to create Xephyr stdout pipe", zap.Error(err))
	}

	xephyrStderr, err := xephyr.StderrPipe()
	if err != nil {
		logger.Fatal("failed to create Xephyr stderr pipe", zap.Error(err))
	}

	if err := xephyr.Start(); err != nil {
		logger.Fatal("failed to start Xephyr", zap.Error(err))
	}

	go logProcessOutput(logger, "stdout", xephyrStdout)
	go logProcessOutput(logger, "stderr", xephyrStderr)

	if err := waitForX(display); err != nil {
		logger.Fatal("failed to wait for X", zap.Error(err))
	}

	logger.Info("X display ready", zap.String("display", display))

	wm := exec.CommandContext(ctx, "./wm")
	wm.Env = append(wm.Environ(), "DISPLAY="+display)

	wm.Stdout = os.Stdout
	wm.Stderr = os.Stderr

	if err := wm.Run(); err != nil {
		logger.Error("WM exited", zap.Error(err))
	}

	if err := xephyr.Wait(); err != nil {
		logger.Debug("Xephyr exited", zap.Error(err))
	}

	logger.Info("development environment stopped")
}

func logProcessOutput(
	logger *zap.Logger,
	source string,
	reader io.Reader,
) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "xkbcomp") ||
			strings.Contains(line, "Could not resolve keysym") ||
			strings.Contains(line, "Errors from xkbcomp are not fatal") {
			continue
		}

		logger.Debug("Xephyr",
			zap.String("source", source),
			zap.String("message", line),
		)
	}

	if err := scanner.Err(); err != nil {
		logger.Error(
			"failed reading Xephyr output",
			zap.String("source", source),
			zap.Error(err),
		)
	}
}

func waitForX(display string) error {
	const timeout = 5 * time.Second
	const retry = 50 * time.Millisecond

	addr := filepath.Join(os.TempDir(), ".X11-unix", "X"+display[1:])

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.Dial("unix", addr)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(retry)
	}

	return fmt.Errorf("timed out waiting for X display %s", display)
}
