#!/usr/bin/env just --justfile

set windows-shell := ["powershell.exe", "-NoLogo", "-Command"]

export CLICOLOR_FORCE := "1"

default:
  just --list

build-backend:
  go build -buildvcs=false -o {{"." / ".cache" / "backend.exe ." / "backend" / "cmd" / "pixelpics" / "."}}

dev:
  ./frontend/node_modules/.bin/concurrently \
    --names "API,WEB" \
    --prefix-colors "bgBlue.bold,bgMagenta.bold" \
    '{{just_executable()}} --justfile {{justfile()}} dev-backend' \
    '{{just_executable()}} --justfile {{justfile()}} dev-frontend'

# TODO: use go tools when go 1.24 comes out
dev-backend:
  go run github.com/air-verse/air \
      --build.cmd '{{just_executable()}} --justfile {{justfile()}} build-backend' \
      --build.bin '{{join(".cache", "backend.exe")}}' \
      --build.exclude_dir 'frontend' \
      -tmp_dir '{{"." / ".cache"}}' serve --dir '{{"." / "pb_data"}}' --dev

dev-frontend:
  cd frontend; bun dev
