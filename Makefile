.DEFAULT_GOAL := build
.PHONY:gen fmt vet build run air
gen:
		go generate .
fmt: gen
		go fmt .
vet: fmt
		go vet .
build: vet
		go build -o ./paste .
run: build
		./paste
air:
		air --build.cmd "make build" --build-bin "./paste"

