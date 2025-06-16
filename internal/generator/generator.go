package generator

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/gera2ld/caddy-gen/internal/config"
	"github.com/gera2ld/caddy-gen/internal/docker"
)

type SiteConfig struct {
	Hostnames       []string
	Port            int
	PathMatcher     string
	Name            string
	HostDirectives  []string
	ProxyDirectives []string
	ProxyIP         string
}

type Generator struct {
	docker *docker.Client
	config *config.Config
}

func NewGenerator(dockerClient *docker.Client, cfg *config.Config) *Generator {
	return &Generator{
		docker: dockerClient,
		config: cfg,
	}
}

func (g *Generator) GenerateConfig() (string, error) {
	containers, err := g.docker.ListContainers()
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %v", err)
	}
	siteConfigs := g.processSiteConfigs(containers)
	groups := g.groupSiteConfigs(siteConfigs)
	return g.generateCaddyConfig(groups), nil
}

func (g *Generator) processSiteConfigs(containers []container.Summary) []SiteConfig {
	var siteConfigs []SiteConfig
	for _, ct := range containers {
		configs := g.processContainer(ct)
		siteConfigs = append(siteConfigs, configs...)
	}
	return siteConfigs
}

func (g *Generator) groupSiteConfigs(siteConfigs []SiteConfig) map[string][]SiteConfig {
	groups := make(map[string][]SiteConfig)
	for _, item := range siteConfigs {
		key := strings.Join(item.Hostnames, " ")
		groups[key] = append(groups[key], item)
	}
	return groups
}

func (g *Generator) generateCaddyConfig(groups map[string][]SiteConfig) string {
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var configParts []string
	for i, hostnames := range keys {
		group := groups[hostnames]
		configParts = append(configParts, g.generateHostConfig(hostnames, group, i))
	}
	return strings.Join(configParts, "\n\n")
}

func (g *Generator) generateHostConfig(hostnames string, group []SiteConfig, index int) string {
	hostMatcher := fmt.Sprintf("@caddy-gen-%d", index)
	var sectionLines []string
	sectionLines = append(sectionLines, fmt.Sprintf("%s host %s", hostMatcher, hostnames))
	sectionLines = append(sectionLines, fmt.Sprintf("handle %s {", hostMatcher))
	sectionLines = append(sectionLines, g.generateDirectives(group, "host")...)
	sectionLines = append(sectionLines, g.generateDirectives(group, "proxy")...)
	sectionLines = append(sectionLines, "}")
	return strings.Join(sectionLines, "\n")
}

func (g *Generator) generateDirectives(group []SiteConfig, directiveType string) []string {
	var lines []string
	for _, item := range group {
		if directiveType == "host" {
			for _, directive := range item.HostDirectives {
				lines = append(lines, fmt.Sprintf("  %s", directive))
			}
		} else {
			lines = append(lines, fmt.Sprintf("  # %s", item.Name))
			lines = append(lines, fmt.Sprintf("  reverse_proxy %s {", item.PathMatcher))
			for _, directive := range item.ProxyDirectives {
				lines = append(lines, fmt.Sprintf("    %s", directive))
			}
			lines = append(lines, fmt.Sprintf("    to %s:%d", item.ProxyIP, item.Port))
			lines = append(lines, "  }")
		}
	}
	return lines
}

func (g *Generator) processContainer(ct container.Summary) []SiteConfig {
	var configs []SiteConfig
	rawBind, exists := ct.Labels["virtual.bind"]
	if !exists || strings.TrimSpace(rawBind) == "" {
		return configs
	}
	var config *SiteConfig = nil
	for _, bindInfo := range strings.Split(rawBind, "\n") {
		bindInfo = strings.TrimSpace(bindInfo)
		if bindInfo == "" || bindInfo[0] == '#' {
			continue
		}
		parts := strings.Fields(bindInfo)
		port, err := strconv.Atoi(parts[0])
		if err == nil {
			var proxyIP string
			if networkSettings, exists := ct.NetworkSettings.Networks[g.config.Network]; exists {
				proxyIP = networkSettings.IPAddress
			}
			configs = append(configs, SiteConfig{
				Name:    strings.TrimPrefix(ct.Names[0], "/"),
				Port:    port,
				ProxyIP: proxyIP,
			})
			config = &configs[len(configs)-1]
			if strings.HasPrefix(parts[1], "/") {
				config.PathMatcher = parts[1]
				config.Hostnames = parts[2:]
			} else {
				config.Hostnames = parts[1:]
			}
			continue
		}
		if config == nil {
			log.Printf("Ignored invalid config: %s\n", bindInfo)
			continue
		}
		g.processDirective(bindInfo, config)
	}
	return configs
}

func (g *Generator) processDirective(directive string, config *SiteConfig) {
	directive = strings.TrimSpace(directive)
	if strings.HasPrefix(directive, "host:") {
		config.HostDirectives = append(config.HostDirectives, strings.TrimSpace(directive[5:]))
	} else {
		config.ProxyDirectives = append(config.ProxyDirectives, directive)
	}
}
