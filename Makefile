VERSION?=v0.1.0
PKG=github.com/nhduc2001kt/hyperv-csi-driver
GIT_COMMIT?=$(shell git rev-parse HEAD)
BUILD_DATE?=$(shell date -u -Iseconds)
LDFLAGS?="-X ${PKG}/pkg/driver.driverVersion=${VERSION} -X ${PKG}/pkg/cloud.driverVersion=${VERSION} -X ${PKG}/pkg/driver.gitCommit=${GIT_COMMIT} -X ${PKG}/pkg/driver.buildDate=${BUILD_DATE} -s -w"

OS?=$(shell go env GOHOSTOS)
ARCH?=$(shell go env GOHOSTARCH)
ifeq ($(OS),windows)
	BINARY=hyperv-csi-driver.exe
	OSVERSION?=ltsc2022
else
	BINARY=hyperv-csi-driver
	OSVERSION?=debian
endif

GO_SOURCES=go.mod go.sum $(shell find pkg cmd -type f -name "*.go")

## Default target
# When no target is supplied, make runs the first target that does not begin with a .
# Alias that to building the binary
.PHONY: default
default: bin/$(BINARY)

## Builds

bin:
	@mkdir -p $@

bin/$(BINARY): $(GO_SOURCES) | bin
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -mod=readonly -ldflags ${LDFLAGS} -o $@ ./cmd/
