# caddy-gen

Generate Caddy configuration for docker sites automatically.

## Usage

Create a Docker network for all services that are reverse-proxied by Caddy, e.g. `gateway` here.

Create a `caddy-gen.yml` from `caddy-gen.example.yml`.

```yaml
services:
  caddy-gen:
    image: gera2ld/caddy-gen
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /caddy/config:/data
    environment:
      - CADDY_GEN_NETWORK=gateway
      - CADDY_GEN_OUTFILE=/data/docker-sites.caddy
      - CADDY_GEN_NOTIFY={"containerId":"caddy","workingDir":"/etc/caddy","command":["caddy","reload"]}
```

Then include the generated `docker-sites.caddy` in your Caddyfile.

```caddy
import path/to/docker-sites.caddy
```

Given a service with the following compose file:

```yaml
services:
  my-service:
    # ...
    labels:
      virtual.bind: 80 my-service.example.com

networks:
  default:
    external: true
    name: gateway
```

Caddy config will be generated automatically to proxy `my-service.example.com` to `my-service:80`.

## Development

This project is written in Go and uses Docker for containerization.

### Building

```bash
# Build the Docker image
just build

# Push the Docker image
just push

# Run the Docker container
just run

# Run locally for development
just dev
```

### Environment Variables

- `CADDY_GEN_NETWORK`: The Docker network to monitor (default: `gateway`)
- `CADDY_GEN_OUTFILE`: The output file for Caddy configuration (default: `docker-sites.caddy`)
- `CADDY_GEN_NOTIFY`: JSON configuration for notifying Caddy to reload (format: `{"containerId":"caddy","workingDir":"/etc/caddy","command":["caddy","reload"]}`)

### Label Format

The `virtual.bind` label supports the following format:

```
PORT [PATH] HOSTNAME1 [HOSTNAME2...]
DIRECTIVE1
DIRECTIVE2
```

- `PATH`: Optional path prefix for the reverse proxy
- `PORT`: The port to proxy to
- `HOSTNAME`: One or more hostnames to match
- `DIRECTIVE`: Optional directives, prefixed with `host:` for host-level directives or without prefix for proxy-level directives
