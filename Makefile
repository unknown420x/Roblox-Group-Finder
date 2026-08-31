APP := groupfinder
VERSION ?= dev

BUILD_DIR := dist

LDFLAGS := -s -w -X main.version=1.0.0

.PHONY: all build release test vet race clean linux windows macos

all: test build

build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(APP) ./cli

linux:
	@mkdir -p $(BUILD_DIR)

	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -trimpath -ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(APP)-linux-amd64 ./cli

	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
		go build -trimpath -ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(APP)-linux-arm64 ./cli

windows:
	@mkdir -p $(BUILD_DIR)

	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
		go build -trimpath -ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(APP)-windows-amd64.exe ./cli

	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 \
		go build -trimpath -ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(APP)-windows-arm64.exe ./cli

macos:
	@mkdir -p $(BUILD_DIR)

	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
		go build -trimpath -ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(APP)-macos-amd64 ./cli

	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
		go build -trimpath -ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(APP)-macos-arm64 ./cli

release: clean test linux windows macos
	@cd $(BUILD_DIR) && sha256sum * > SHA256SUMS
	@echo "Release builds created in $(BUILD_DIR)/"

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
