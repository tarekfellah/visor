MAJOR   := 0
MINOR   := 5
PATCH   := 9
VERSION := $(MAJOR).$(MINOR).$(PATCH)
GOPATH  ?= $(PWD)
GOBIN   ?= $(GOPATH)/bin
GOFLAGS := -ldflags "-X main.VERSION $(VERSION)"

# LOCAL #

default:
	@GOPATH=$(GOPATH) go get $(GOFLAGS) -d ./cmd/visor
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install $(GOFLAGS) ./cmd/visor
	@echo "built $(GOBIN)/visor v$(VERSION)"

# DEBIAN PACKAGING #

DEB_NAME=visor
DEB_URL=http://github.com/soundcloud/visor
DEB_VERSION=$(VERSION)
DEB_DESCRIPTION=A command line interface for visor
DEB_MAINTAINER=Daniel Bornkessel <daniel@soundcloud.com>

include deb.mk

debroot:
	GOBIN=$(DEB_ROOT)/usr/bin $(MAKE)

# BUILD #

build: clean debroot debbuild

clean: debclean
	go clean
	rm -rf bin src pkg
