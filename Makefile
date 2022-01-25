BUILD_DATE := `date -u +%Y%m%d`
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo v0.0.1)
GIT_SHA := $(shell git rev-parse HEAD)
APP_NAME := prettify
PROJECT := github.com/gsmcwhirter/$(APP_NAME)

GOPROXY ?= https://proxy.golang.org

# can specify V=1 on the line with `make` to get verbose output
V ?= 0
Q = $(if $(filter 1,$V),,@)

.DEFAULT_GOAL := help

all: debug  ## Download dependencies and do a debug build

build-debug: version
	$Q go build -v -ldflags "-X main.AppName=prettify -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/prettify -race $(PROJECT)/cmd/prettify
	$Q go build -v -ldflags "-X main.AppName=logfollow -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/logfollow -race $(PROJECT)/cmd/logfollow

build-release-osx: version
	$Q go build -v -ldflags "-s -w -X main.AppName=prettify -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/prettify-osx $(PROJECT)/cmd/prettify
	$Q go build -v -ldflags "-s -w -X main.AppName=logfollow -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/logfollow-osx $(PROJECT)/cmd/logfollow

build-release-linux: version
	$Q GOOS=linux go build -v -ldflags "-s -w -X main.AppName=prettify -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/prettify-linux $(PROJECT)/cmd/prettify
	$Q GOOS=linux go build -v -ldflags "-s -w -X main.AppName=logfollow -X main.BuildVersion=$(VERSION) -X main.BuildSHA=$(GIT_SHA) -X main.BuildDate=$(BUILD_DATE)" -o bin/logfollow-linux $(PROJECT)/cmd/logfollow

debug: vet generate build-debug  ## Debug build: create a dev build (enable race detection, don't strip symbols)

deps:  ## download dependencies
	$Q GOPROXY=$(GOPROXY) go mod download

generate:  ## run a go generate
	$Q GOPROXY=$(GOPROXY) go generate ./...

test:  ## run go test
	$Q GOPROXY=$(GOPROXY) go test -cover ./...

version:  ## Print the version string and git sha that would be recorded if a release was built now
	$Q echo $(VERSION)
	$Q echo $(GIT_SHA)

vet: deps generate ## run various linters and vetters
	$Q bash -c 'for d in $$(go list -f {{.Dir}} ./...); do gofmt -s -w $$d/*.go; done'
	$Q bash -c 'for d in $$(go list -f {{.Dir}} ./...); do goimports -w -local $(PROJECT) $$d/*.go; done'
	$Q golangci-lint run -E revive,gosimple,staticcheck ./...
	$Q golangci-lint run -E deadcode,depguard,errcheck,gocritic,gofmt,goimports,gosec,govet,ineffassign,nakedret,prealloc,structcheck,typecheck,unconvert,varcheck ./...

help:  ## Show the help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' ./Makefile
