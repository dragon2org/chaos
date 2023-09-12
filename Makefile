PACKAGE=github.com/dragon2org/chaos
PREFIX=$(shell pwd)
CMD_PACKAGE=${PACKAGE}/chaos
OUTPUT_DIR=${PREFIX}/bin
OUTPUT_FILE=${OUTPUT_DIR}/chaos
COMMIT_ID=$(shell git rev-parse --short HEAD)
VERSION=$(shell git describe --tags || echo "v0.0.0")
VERSION_IMPORT_PATH=main
BUILD_TIME=$(shell date '+%Y-%m-%dT%H:%M:%S%Z')
VCS_BRANCH=$(shell git symbolic-ref --short -q HEAD)

# build args
BUILD_ARGS := \
    -ldflags "-X $(VERSION_IMPORT_PATH).appName=$(PACKAGE) \
    -X $(VERSION_IMPORT_PATH).version=$(VERSION) \
    -X $(VERSION_IMPORT_PATH).revision=$(COMMIT_ID) \
    -X $(VERSION_IMPORT_PATH).branch=$(VCS_BRANCH) \
    -X $(VERSION_IMPORT_PATH).buildDate=$(BUILD_TIME)"
EXTRA_BUILD_ARGS=

export CGO_ENABLED=0

.PONY: all

all: lint test build

lint:
	@echo "+ $@"
	golangci-lint run --timeout 3m ./...

test:
	@echo "+ $@"
	go test $(BUILD_ARGS) $(EXTRA_BUILD_ARGS)  -covermode=count -coverprofile=count.out ./...

coverage:
	@echo "+ $@"
	@if [ ! -f "count.out" ]; then make test ; fi
	go tool cover -func=count.out

build-linux-amd64:
	@echo "+ $@"
	GOOS=linux GOARCH=amd64 go build $(BUILD_ARGS) $(EXTRA_BUILD_ARGS) -o ${OUTPUT_FILE}-linux-amd64 $(CMD_PACKAGE)

build-win-amd64:
	@echo "+ $@"
	GOOS=windows GOARCH=amd64 go build $(BUILD_ARGS) $(EXTRA_BUILD_ARGS) -o ${OUTPUT_FILE}-win-amd64 $(CMD_PACKAGE)

build-darwin-amd64:
	@echo "+ $@"
	GOOS=darwin GOARCH=amd64 go build $(BUILD_ARGS) $(EXTRA_BUILD_ARGS) -o ${OUTPUT_FILE}-darwin-amd64 $(CMD_PACKAGE)

build: build-linux-amd64 build-win-amd64
	@echo "+ $@"

install: build
	cp bin/chaos ${HOME}/.local/bin/chaos

clean:
	@echo "+ $@"
	@if [ -d "bin/" ]; then rm -rf bin/ ; fi
	@if [ -f "count.out" ]; then rm count.out; fi
