package service

import (
	"log"
	"os"

	"github.com/gera2ld/caddy-gen/internal/config"
	"github.com/gera2ld/caddy-gen/internal/docker"
	"github.com/gera2ld/caddy-gen/internal/generator"
)

// Service is the main service
type Service struct {
	docker    *docker.Client
	generator *generator.Generator
	config    *config.Config
}

// NewService creates a new Service
func NewService() (*Service, error) {
	cfg := config.NewConfig()
	dockerClient, err := docker.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	gen := generator.NewGenerator(dockerClient, cfg)
	return &Service{
		docker:    dockerClient,
		generator: gen,
		config:    cfg,
	}, nil
}

// Close closes the service
func (s *Service) Close() error {
	return s.docker.Close()
}

// Run runs the service
func (s *Service) Run() error {
	s.CheckConfig()
	log.Println("Waiting for Docker events...")
	s.docker.WatchEvents(s.CheckConfig)
	return nil
}

// CheckConfig checks and updates the configuration
func (s *Service) CheckConfig() {
	currentConfig := s.readConfig(s.config.OutFile)
	newConfig, err := s.generator.GenerateConfig()
	if err != nil {
		log.Printf("Failed to generate config: %v", err)
		return
	}
	if currentConfig != newConfig {
		s.writeConfig(s.config.OutFile, newConfig)
		s.notifyConfigChange()
	} else {
		log.Println("No change, skip notifying")
	}
}

// readConfig reads the configuration from the file
func (s *Service) readConfig(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Failed to read config file: %v", err)
	}
	return string(data)
}

// writeConfig writes the configuration to the file
func (s *Service) writeConfig(filePath, config string) {
	err := os.WriteFile(filePath, []byte(config), 0644)
	if err != nil {
		log.Printf("Failed to write config: %v", err)
	}
	log.Printf("Caddy config written: %s", filePath)
}

// notifyConfigChange notifies that the configuration has changed
func (s *Service) notifyConfigChange() {
	s.docker.Notify()
}
