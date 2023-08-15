.PHONY: build
build:
	mkdir -p bin/
	GOARCH=wasm GOOS=js go build -o bin/office.wasm cmd/office/main.go
	go build -o bin/serve cmd/wasm-serve/main.go
