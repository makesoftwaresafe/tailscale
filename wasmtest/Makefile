run:
	cp "${GOROOT}/misc/wasm/wasm_exec.js" ./
	GOOS=js GOARCH=wasm ../tool/go build -o test.wasm ./wasmmod/
	go run serve.go
