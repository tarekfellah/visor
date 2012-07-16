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

build: package

########## packaging
FPM_EXECUTABLE:=$$(dirname $$(dirname $$(gem which fpm)))/bin/fpm
FPM_ARGS=-t deb -m 'Visor authors (see page), Daniel Bornkessel <daniel@soundcloud.com> (packaging)' --url http://github.com/soundcloud/visor -s dir
FAKEROOT=fakeroot
RELEASE=$$(cat .release 2>/dev/null || echo "0")

package: fmt update_version goclean gobuild bump_package_release
	- mkdir -p $(FAKEROOT)/usr/bin
	cp bin/visor $(FAKEROOT)/usr/bin
	-rm *.deb

	$(FPM_EXECUTABLE) -n "visor" \
		-C $(FAKEROOT) \
		--description "visor cli" \
		$(FPM_ARGS) -t deb -v $$(cat VERSION) --iteration $(RELEASE) .;


bump_package_release:
		echo $$(( $(RELEASE) + 1 )) > .release

clean: goclean
	rm -rf bin $(FAKEROOT)
