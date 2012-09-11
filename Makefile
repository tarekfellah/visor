VERSION := $$(cat VERSION)

GOPATH   := $(PWD)
GOBIN    ?= $(GOPATH)/bin
GOFLAGS  := -ldflags "-X main.VERSION $(VERSION)"
PKG_PATH := $(GOPATH)/src/github.com/soundcloud/visor

# LOCAL #
default: $(PKG_PATH)
	GOPATH=$(GOPATH) go get $(GOFLAGS) -d ./cmd/visor
	GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install $(GOFLAGS) ./cmd/visor
	echo "built $(GOBIN)/visor v$(VERSION)"

$(PKG_PATH):
	mkdir -p $$(dirname $(PKG_PATH))
	ln -sf $(PWD) $(PKG_PATH)

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
