package generator

import (
	"fmt"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/gera2ld/caddy-gen/internal/config"
	"github.com/gera2ld/caddy-gen/internal/docker"
)

func TestParseBindInfo(t *testing.T) {
	// Create test container
	container := container.Summary{
		Names:  []string{"/test-container"},
		Labels: map[string]string{},
		NetworkSettings: &container.NetworkSettingsSummary{
			Networks: map[string]*network.EndpointSettings{
				"gateway": {
					IPAddress: "172.17.0.2",
				},
			},
		},
	}

	// Create generator
	cfg := &config.Config{Network: "gateway"}
	dockerClient := &docker.Client{} // Mock client
	generator := NewGenerator(dockerClient, cfg)

	// Test simple bind
	container.Labels["virtual.bind"] = "80 example.com"
	configs, err := generator.processContainer(container)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if len(configs) != 1 {
		t.Errorf("configs = %v; want exactly 1 config", configs)
	}
	siteConfig := configs[0]
	if siteConfig.Hostnames[0] != "example.com" {
		t.Errorf("siteConfig.Hostnames = %v; want [example.com]", siteConfig.Hostnames)
	}
	if siteConfig.Port != 80 {
		t.Errorf("siteConfig.Port = %d; want 80", siteConfig.Port)
	}
	if siteConfig.PathMatcher != "" {
		t.Errorf("siteConfig.PathMatcher = %s; want \"\"", siteConfig.PathMatcher)
	}
	if siteConfig.Name != "test-container" {
		t.Errorf("siteConfig.Name = %s; want test-container", siteConfig.Name)
	}
	if siteConfig.ProxyIP != "172.17.0.2" {
		t.Errorf("siteConfig.ProxyIP = %s; want 172.17.0.2", siteConfig.ProxyIP)
	}

	// Test bind with path
	container.Labels["virtual.bind"] = "80 /api example.com"
	configs, err = generator.processContainer(container)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if len(configs) != 1 {
		t.Errorf("configs = %v; want exactly 1 config", configs)
	}
	siteConfig = configs[0]
	if siteConfig.PathMatcher != "/api" {
		t.Errorf("siteConfig.PathMatcher = %s; want /api", siteConfig.PathMatcher)
	}

	// Test bind with directives
	container.Labels["virtual.bind"] = `80 example.com
header Server "My Server"
host:tls {
internal
}`
	configs, err = generator.processContainer(container)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if len(configs) != 1 {
		t.Errorf("configs = %v; want exactly 1 config", configs)
	}
	siteConfig = configs[0]
	fmt.Printf("%v\n", siteConfig.HostDirectives[0])
	if len(siteConfig.HostDirectives) != 1 || siteConfig.HostDirectives[0] != "tls {\ninternal\n}" {
		t.Errorf("siteConfig.HostDirectives = %v; want [tls internal]", siteConfig.HostDirectives)
	}
	if len(siteConfig.ProxyDirectives) != 1 || siteConfig.ProxyDirectives[0] != "header Server \"My Server\"" {
		t.Errorf("siteConfig.ProxyDirectives = %v; want [header Server \"My Server\"]", siteConfig.ProxyDirectives)
	}

	// Test invalid bind
	container.Labels["virtual.bind"] = "Invalid"
	configs, err = generator.processContainer(container)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if len(configs) != 0 {
		t.Errorf("configs = %v; want no config", configs)
	}
}

func TestProcessContainer(t *testing.T) {
	// Create test container
	container := container.Summary{
		Names: []string{"/test-container"},
		Labels: map[string]string{
			"virtual.bind": `80 example.com
8080 /api api.example.com`,
		},
		NetworkSettings: &container.NetworkSettingsSummary{
			Networks: map[string]*network.EndpointSettings{
				"gateway": {
					IPAddress: "172.17.0.2",
				},
			},
		},
	}

	// Create generator
	cfg := &config.Config{Network: "gateway"}
	dockerClient := &docker.Client{} // Mock client
	generator := NewGenerator(dockerClient, cfg)

	// Test process container
	configs, err := generator.processContainer(container)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if len(configs) != 2 {
		t.Fatalf("processContainer() returned %d configs; want 2", len(configs))
	}

	// Check first config
	if configs[0].Port != 80 || configs[0].Hostnames[0] != "example.com" {
		t.Errorf("configs[0] = %+v; want Port=80, Hostnames=[example.com]", configs[0])
	}

	// Check second config
	if configs[1].Port != 8080 || configs[1].Hostnames[0] != "api.example.com" || configs[1].PathMatcher != "/api" {
		t.Errorf("configs[1] = %+v; want Port=8080, Hostnames=[api.example.com], PathMatcher=/api", configs[1])
	}
}

