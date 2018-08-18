BUILD_DATE := `date -u +%Y%m%d`
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo v0.0.1)
GIT_SHA := $(shell git rev-parse HEAD)
APP_NAME := prettify
PROJECT := github.com/gsmcwhirter/$(APP_NAME)
SERVER := evogames.org:~/bin/

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

release: vet generate test build-release-bundles  ## Release build: create a release build (disable race detection, strip symbols)

release-upload: release upload-release-bundles  ## Release build+upload: create a release build and distribute release files to s3

generate:
	$Q go generate ./...

test:  ## Run the tests
	$Q go test -cover ./...

upload:
	$Q scp  ./bin/$(APP_NAME).gz ./bin/$(APP_NAME)-$(VERSION).gz $(SERVER)

version:  ## Print the version string and git sha that would be recorded if a release was built now
	$Q echo $(VERSION)
	$Q echo $(GIT_SHA)

vet:  ## Run the linter
	$Q golint ./...
	$Q go vet ./...
	$Q gometalinter -D gas -D gocyclo -D goconst -e .pb.go -e _easyjson.go --warn-unmatched-nolint --enable-gc --deadline 180s ./...

help:  ## Show the help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' ./Makefile
