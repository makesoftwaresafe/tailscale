run:
	cp "${GOROOT}/misc/wasm/wasm_exec.js" ./
	GOOS=js GOARCH=wasm ../tool/go build \
		-tags tailscale_go,osusergo,netgo,nethttpomithttp2,omitidna,omitpemdecrypt \
		-o test.wasm \
		./wasmmod/
	../../corp/tool/yarn
	go run serve.go

run-prod:
	cp "${GOROOT}/misc/wasm/wasm_exec.js" ./
	GOOS=js GOARCH=wasm ../tool/go build \
		-tags tailscale_go,osusergo,netgo,nethttpomithttp2,omitidna,omitpemdecrypt \
		-trimpath \
		-ldflags "-s -w" \
		-o test.wasm \
		./wasmmod/
	../../corp/tool/yarn
	go run serve.go --dev=0
