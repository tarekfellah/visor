bin/visor: gofmt update_version
	mkdir -p bin
	go build
	go build -o bin/visor ./visor

install: update_version gobuild
	mkdir -p $${DESTDIR-/usr/local}/bin
	cp bin/visor $${DESTDIR-/usr/local}/bin

update_version:
	grep "const VERSION_STRING = \"v$$(cat VERSION)\"" visor/main.go || sed -i -e "s/const VERSION_STRING .*/const VERSION_STRING = \"$$(cat VERSION)\"/" visor/main.go
	grep ".*version '$$(cat VERSION)'" visor.rb || sed -i -e "s/.*version '[\.0-9]*'$$/  version '$$(cat VERSION)'/" visor.rb


########### local build:
PKG=github.com/soundcloud/visor
SUB_PKG=github.com/soundcloud/visor/visor
include go.mk

build: clean debroot debbuild

########## packaging
DEB_NAME=visor
DEB_URL=http://github.com/soundcloud/visor
DEB_VERSION=$$(cat VERSION)
DEB_DESCRIPTION=A command line interface for visor
DEB_MAINTAINER=Daniel Bornkessel <daniel@soundcloud.com>

include deb.mk

debroot:
	DESTDIR=$(DEB_ROOT) $(MAKE) install

clean: goclean debclean
	rm -rf bin $(DEB_ROOT)
