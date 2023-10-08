.PHONY: build
build:
	mkdir -p bin/
	GOOS=js GOARCH=wasm go build -o bin/hashira.wasm cmd/hashira/main.go
	go build -o bin/serve cmd/wasm-serve/main.go

.PHONY: release
release:
	mkdir -p releases/${TAG}
	GOOS=js GOARCH=wasm go build -o bin/hashira.wasm cmd/hashira/main.go
	mv bin/hashira.wasm releases/${TAG}/hashira.wasm
	cp ui/hashira.js releases/${TAG}/hashira.js
	cd releases/${TAG} && zip -r -9 ${TAG}.zip *.js *.wasm
	mv releases/${TAG}/${TAG}.zip releases/${TAG}.zip
	cd releases/${TAG} && tar -cvzf ${TAG}.tar.gz *.js *.wasm
	mv releases/${TAG}/${TAG}.tar.gz releases/${TAG}.tar.gz
	rm -rf releases/${TAG}