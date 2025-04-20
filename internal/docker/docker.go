package docker

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/gera2ld/caddy-gen/internal/config"
)

// Client wraps the Docker client with additional functionality
type Client struct {
	client *client.Client
	config *config.Config
}

// NewClient creates a new Docker client
func NewClient(cfg *config.Config) (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}
	return &Client{
		client: cli,
		config: cfg,
	}, nil
}

// Close closes the Docker client
func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) ListContainers() ([]container.Summary, error) {
	ctx := context.Background()
	args := c.createListFilter()
	return c.client.ContainerList(ctx, container.ListOptions{
		Filters: args,
	})
}

func (c *Client) createListFilter() filters.Args {
	args := filters.NewArgs()
	args.Add("network", c.config.Network)
	args.Add("status", "created")
	args.Add("status", "running")
	return args
}

// Notify notifies the Caddy container to reload
func (c *Client) Notify() {
	if c.config.Notify == nil {
		return
	}
	if c.config.Notify.ContainerID == "" {
		name := c.config.Notify.Command[0]
		args := c.config.Notify.Command[1:]
		cmd := exec.Command(name, args...)
		cmd.Run()
		return
	}
	log.Printf("Notify: %+v", c.config.Notify)
	ctx := context.Background()
	c.executeCommand(ctx, c.config.Notify)
}

func (c *Client) executeCommand(ctx context.Context, notifyConfig *config.NotifyConfig) {
	execConfig := c.createExecConfig(notifyConfig)
	resp, err := c.client.ContainerExecCreate(ctx, notifyConfig.ContainerID, execConfig)
	if err != nil {
		log.Printf("Failed to create exec: %v", err)
		return
	}
	err = c.client.ContainerExecStart(ctx, resp.ID, container.ExecStartOptions{})
	if err != nil {
		log.Printf("Failed to start exec: %v", err)
	}
}

func (c *Client) createExecConfig(notifyConfig *config.NotifyConfig) container.ExecOptions {
	return container.ExecOptions{
		Cmd:          notifyConfig.Command,
		WorkingDir:   notifyConfig.WorkingDir,
		AttachStdout: false,
		AttachStderr: false,
		Detach:       true,
	}
}

// WatchEvents watches for Docker events and calls the callback function
func (c *Client) WatchEvents(callback func()) {
	ctx := context.Background()
	args := c.createEventFilter()
	debouncedCallback := debounce(callback, 1*time.Second)
	c.watchEventLoop(ctx, args, debouncedCallback)
}

// createEventFilter creates a filter for container events
func (c *Client) createEventFilter() filters.Args {
	args := filters.NewArgs()
	args.Add("type", "container")
	args.Add("event", "start")
	args.Add("event", "stop")
	return args
}

// watchEventLoop watches for Docker events in a loop
func (c *Client) watchEventLoop(ctx context.Context, args filters.Args, callback func()) {
	for {
		messages, errs := c.client.Events(ctx, events.ListOptions{
			Filters: args,
		})
		c.processEvents(messages, errs, callback)
	}
}

// processEvents processes Docker events
func (c *Client) processEvents(messages <-chan events.Message, errs <-chan error, callback func()) {
	for {
		select {
		case <-messages:
			callback()
		case err := <-errs:
			if err != nil {
				log.Printf("Error receiving events: %v", err)
				time.Sleep(5 * time.Second) // Wait before reconnecting
				return
			}
		}
	}
}

// Debounce function to avoid multiple callbacks
func debounce(f func(), delay time.Duration) func() {
	var timer *time.Timer
	return func() {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(delay, f)
	}
}
