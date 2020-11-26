VERSION=0.0.3
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"
GO111MODULE=on

all: mackerel-plugin-httping

.PHONY: mackerel-plugin-httping

mackerel-plugin-httping: main.go
	go build $(LDFLAGS) -o mackerel-plugin-httping

linux: main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o mackerel-plugin-httping

deps:
	go get -d
	go mod tidy

deps-update:
	go get -u -d
	go mod tidy

clean:
	rm -rf mackerel-plugin-httping

check:
	go test ./...

tag:
	git tag v${VERSION}
	git push origin v${VERSION}
	git push origin master
