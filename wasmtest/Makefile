run:
	cp "${GOROOT}/misc/wasm/wasm_exec.js" ./
	GOOS=js GOARCH=wasm ../tool/go build -o test.wasm ./wasmmod/
	../../corp/tool/yarn
	go run serve.go
