VERSION=$$(cat VERSION)

PKG=github.com/soundcloud/visor
PKG_GOPATH=$(PWD)/src/$(PKG)
GOFLAGS=-v -x -ldflags "-X main.VERSION_STRING $(VERSION)"
GOPATH?=$(PWD)

compile: $(PKG_GOPATH)
	GOPATH=$(GOPATH) go get $(GOFLAGS) -d $(PKG)/visor
	GOPATH=$(GOPATH) go install $(GOFLAGS) $(PKG)/visor

$(PKG_GOPATH):
	mkdir -p $$(dirname $(PKG_GOPATH))
	ln -sfn $(PWD) $(PKG_GOPATH)

########## packaging

DEB_NAME=visor
DEB_URL=http://github.com/soundcloud/visor
DEB_VERSION=$(VERSION)
DEB_DESCRIPTION=A command line interface for visor
DEB_MAINTAINER=Daniel Bornkessel <daniel@soundcloud.com>

include deb.mk

debroot:
	GOBIN=$(DEB_ROOT)/usr/bin $(MAKE)

##########

build: clean debroot debbuild

clean: debclean
	rm -rf bin src pkg $(DEB_ROOT)
