default:
  just --list

build:
  CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o caddy-gen .

dev:
  go run main.go
