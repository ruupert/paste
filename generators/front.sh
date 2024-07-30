#!/bin/bash
GOOS=js GOARCH=wasm go build -o ./public/out.wasm ./frontend/main.go
