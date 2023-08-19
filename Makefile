.PHONY: build
build:
	mkdir -p bin/
	GOOS=js GOARCH=wasm go build -o bin/hashira.wasm cmd/hashira/main.go
	go build -o bin/serve cmd/wasm-serve/main.go
