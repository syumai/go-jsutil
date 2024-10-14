SHELL=/bin/bash

.PHONY: test
test:
	@PATH=$$(go env GOROOT)/misc/wasm:$$PATH GOOS=js GOARCH=wasm go test -v ./...
