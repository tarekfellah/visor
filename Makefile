LOCAL_GOPATH=${PWD}/local_go_path
FPM_EXECUTABLE:=$$(dirname $$(dirname $$(gem which fpm)))/bin/fpm
FPM_ARGS=-t deb -m 'Visor authors (see page), Daniel Bornkessel <daniel@soundcloud.com> (packaging)' --url http://github.com/soundcloud/visor -s dir
FAKEROOT=fakeroot

RELEASE=$$(cat .release 2>/dev/null || echo "0")

bump_package_release:
		echo $$(( $(RELEASE) + 1 )) > .release

compile: fmt
	go build
	go build -o bin/visor ./cmd

$(LOCAL_GOPATH)/src:
	mkdir -p $(LOCAL_GOPATH)/src

$(LOCAL_GOPATH)/src/github.com/soundcloud/doozer: $(LOCAL_GOPATH)/src
	GOPATH=$(LOCAL_GOPATH) go get github.com/soundcloud/doozer

$(LOCAL_GOPATH)/src/github.com/kesselborn/go-getopt: $(LOCAL_GOPATH)/src
	GOPATH=$(LOCAL_GOPATH) go get github.com/kesselborn/go-getopt

local_build:
	GOPATH=$(LOCAL_GOPATH) go build
	GOPATH=$(LOCAL_GOPATH) go build -o bin/visor ./cmd

package: local_build
	- mkdir -p $(FAKEROOT)/usr/bin
	cp bin/visor $(FAKEROOT)/usr/bin

	$(FPM_EXECUTABLE) -n "visor" \
		-C $(FAKEROOT) \
		--description "visor cli" \
		$(FPM_ARGS) -t deb -v $$(cat VERSION) --iteration $(RELEASE) .;

build: fmt $(LOCAL_GOPATH)/src/github.com/soundcloud/doozer $(LOCAL_GOPATH)/src/github.com/kesselborn/go-getopt package bump_package_release
	echo ".git" > .pkgignore
	find . -mindepth 1 -maxdepth 1 | grep -v "\.deb" | sed 's/\.\///g' >> .pkgignore


fmt:
	go fmt ./...

clean:
	git clean -xdf
