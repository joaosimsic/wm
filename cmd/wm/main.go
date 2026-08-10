package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	wmcore "wm/pkg/wm"
)

func main() {
	wm, err := wmcore.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start wm: %v\n", err)
		os.Exit(1)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		wm.Stop()
	}()

	if err := wm.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "wm error: %v\n", err)
		os.Exit(1)
	}
}
