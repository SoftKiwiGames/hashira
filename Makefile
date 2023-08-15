.PHONY: build
build:
	mkdir -p bin/
	GOOS=js GOARCH=wasm go build -o bin/office.wasm cmd/office/main.go
	go build -o bin/serve cmd/wasm-serve/main.go
