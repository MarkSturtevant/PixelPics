#!/usr/bin/env just --justfile

default:
  @just --list

dev:
  ./frontend/node_modules/.bin/concurrently \
    --names "API,WEB" \
    --prefix-colors "bgBlue.bold,bgMagenta.bold" \
    'CLICOLOR_FORCE=1 {{just_executable()}} --justfile {{justfile()}} dev-backend' \
    '{{just_executable()}} --justfile {{justfile()}} dev-frontend'

# TODO: use go tools when go 1.24 comes out
dev-backend:
  go run -modfile=./backend/tools/go.mod github.com/air-verse/air \
      --build.cmd "go build -o ./.cache/backend ./backend/cmd/pixelpics/." \
      --build.bin "./.cache/backend" \
      --build.exclude_dir "frontend" \
      -tmp_dir ./.cache serve

dev-frontend:
  cd frontend && bun dev