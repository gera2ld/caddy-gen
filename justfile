default:
  just --list

build:
  go run build.go

dev:
  go run main.go

test:
  go test ./internal/...
