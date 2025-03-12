default: just --list

build: docker build -t gera2ld/caddy-gen:latest .

push: docker push gera2ld/caddy-gen:latest

run: docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v $(pwd):/data -e CADDY_GEN_OUTFILE=/data/docker-sites.caddy gera2ld/caddy-gen:latest

dev: go run main.go
