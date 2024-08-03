#!/bin/bash
GOOS=js GOARCH=wasm go build -o ./assets/out.wasm ./frontend/main.go
