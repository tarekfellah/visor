bin/visor: bin fmt update_version
	go build
	go build -o bin/visor ./visor

install: gobuild
	mkdir -p $${DESTDIR-/usr/local}/bin
	cp bin/visor $${DESTDIR-/usr/local}/bin

fmt:
	go fmt ./...

update_version:
	grep "const VERSION_STRING = \"v$$(cat VERSION)\"" visor/main.go || sed -i -e "s/const VERSION_STRING .*/const VERSION_STRING = \"v$$(cat VERSION)\"/" visor/main.go
	grep ".*version '$$(cat VERSION)'" visor.rb || sed -i -e "s/.*version '[\.0-9]*'$$/  version '$$(cat VERSION)'/" visor.rb


########### local build:
PKG=github.com/soundcloud/visor
SUB_PKG=github.com/soundcloud/visor/visor
include go.mk

build: clean update_version gobuild debroot debbuild

########## packaging
DEB_NAME=visor
DEB_URL=http://github.com/soundcloud/visor
DEB_VERSION=$$(cat VERSION)
DEB_DESCRIPTION=visor cli
DEB_MAINTAINER=Daniel Bornkessel <daniel@soundcloud.com>

include deb.mk

debroot:
	mkdir -p $(DEB_ROOT)/usr/bin
	cp bin/visor $(DEB_ROOT)/usr/bin

clean: goclean debclean
	rm -rf bin $(DEB_ROOT)
