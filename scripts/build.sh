#!/bin/bash
GOOS=wasip1 GOARCH=wasm go build -o ../dist/plugin-morphe-pydantic-types-v1.0.0.wasm ../cmd/plugin/main.go
