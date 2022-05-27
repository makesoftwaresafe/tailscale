run:
	cp "${GOROOT}/misc/wasm/wasm_exec.js" ./
	GOOS=js GOARCH=wasm ../tool/go build \
		-tags nethttpomithttp2,omitidna,omitpemdecrypt \
		-trimpath \
		-ldflags "-s -w" \
		-o test.wasm \
		./wasmmod/
	../../corp/tool/yarn
	go run serve.go
