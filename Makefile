BINARY := ta-time-source
BINARY_VERSION := v1.0.0
BUILD_DIR := ./build

.PHONY: clean-build clean-all

all: clean-all compile-all

init:
	go mod tidy
	mkdir -p $(BUILD_DIR)

compile-linux: init
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '-s -w' \
		-o $(BUILD_DIR)/$(BINARY)-$(BINARY_VERSION)-linux-amd64 
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '-s -w' \
		-o $(BUILD_DIR)/$(BINARY)-$(BINARY_VERSION)-linux-arm64
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -ldflags '-s -w' \
		-o $(BUILD_DIR)/$(BINARY)-$(BINARY_VERSION)-linux-arm

compile-all: compile-linux

clean-build:
	rm -rf $(BUILD_DIR)

clean-all: clean-build
