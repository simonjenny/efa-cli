build: builds/efa

builds/efa: go.mod $(shell find . -name '*.go' -not -path './builds/*')
	CGO_ENABLED=0 go build -trimpath -o builds/efa ./cmd/efa

build-static: build

release:
	mkdir -p release
	cp builds/efa release/efa

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l .

clean:
	rm -rf builds release static

.PHONY: build build-static release test vet fmt clean