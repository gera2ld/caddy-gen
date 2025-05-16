package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gera2ld/caddy-gen/internal/service"
)

func main() {
	// Create service
	svc, err := service.NewService()
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}
	defer svc.Close()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Run service
	go func() {
		if err := svc.Run(); err != nil {
			log.Fatalf("Service error: %v", err)
		}
	}()

	// Wait for signal
	sig := <-sigCh
	log.Printf("Received signal: %v, shutting down...", sig)
} 