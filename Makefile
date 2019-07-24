BUILD_DATE := `date -u +%Y%m%d`
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo v0.0.1)
GIT_SHA := $(shell git rev-parse HEAD)
APP_NAME := prettify
PROJECT := github.com/gsmcwhirter/$(APP_NAME)
SERVER := evogames.org:~/bin/

GOPROXY ?= https://proxy.golang.org

# can specify V=1 on the line with `make` to get verbose output
V ?= 0
Q = $(if $(filter 1,$V),,@)

.DEFAULT_GOAL := help

all: debug  ## Download dependencies and do a debug build

build-debug: version
	$Q go build -v -ldflags "-X main.AppName=$(APP_NAME) -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(APP_NAME) -race $(PROJECT)/cmd/$(APP_NAME)

build-release: version
	$Q GOOS=linux go build -v -ldflags "-s -w -X main.AppName=$(APP_NAME) -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(APP_NAME) $(PROJECT)/cmd/$(APP_NAME)

build-release-bundles: build-release
	$Q gzip -k -f bin/$(APP_NAME)
	$Q cp bin/$(APP_NAME).gz bin/$(APP_NAME)-$(VERSION).gz

clean:  ## Remove compiled artifacts
	$Q rm bin/*

debug: vet generate build-debug  ## Debug build: create a dev build (enable race detection, don't strip symbols)

deps:  ## download dependencies
	$Q GOPROXY=$(GOPROXY) go mod download

release: vet generate test build-release-bundles  ## Release build: create a release build (disable race detection, strip symbols)

release-upload: release upload  ## Release build+upload: create a release build and distribute release files to s3

generate:  ## run a go generate
	$Q GOPROXY=$(GOPROXY) go generate ./...

test:  ## run go test
	$Q GOPROXY=$(GOPROXY) go test -cover ./...

upload:
	$Q scp  ./bin/$(APP_NAME).gz ./bin/$(APP_NAME)-$(VERSION).gz $(SERVER)

version:  ## Print the version string and git sha that would be recorded if a release was built now
	$Q echo $(VERSION)
	$Q echo $(GIT_SHA)

vet: deps generate ## run various linters and vetters
	$Q bash -c 'for d in $$(go list -f {{.Dir}} ./...); do gofmt -s -w $$d/*.go; done'
	$Q bash -c 'for d in $$(go list -f {{.Dir}} ./...); do goimports -w -local $(PROJECT) $$d/*.go; done'
	$Q golangci-lint run -E golint,gosimple,staticcheck ./...
	$Q golangci-lint run -E deadcode,depguard,errcheck,gocritic,gofmt,goimports,gosec,govet,ineffassign,nakedret,prealloc,structcheck,typecheck,unconvert,varcheck ./...

help:  ## Show the help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' ./Makefile
