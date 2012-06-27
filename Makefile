bin/visor: bin fmt update_version
	go build
	go build -o bin/visor ./cmd

bin:
	mkdir -p bin

install: local_build
	mkdir -p $${DESTDIR-/usr/local}/bin
	cp bin/visor $${DESTDIR-/usr/local}/bin

fmt:
	go fmt ./...

update_version:
	grep "const VERSION_STRING = \"v$$(cat VERSION)\"" cmd/visorcli.go || sed -i -e "s/const VERSION_STRING .*/const VERSION_STRING = \"v$$(cat VERSION)\"/" cmd/visorcli.go
	grep ".*version '$$(cat VERSION)'" visor.rb || sed -i -e "s/.*version '[\.0-9]*'$$/  version '$$(cat VERSION)'/" visor.rb

clean:
	git clean -xdf

########### local build:

LOCAL_GOPATH=${PWD}/.go_path
VISOR_GO_PATH=$(LOCAL_GOPATH)/src/github.com/soundcloud/visor

unexport GIT_DIR
unexport GOROOT
unexport GOBIN

build: fmt package bump_package_release
	echo ".git" > .pkgignore
	find . -mindepth 1 -maxdepth 1 | grep -v "\.deb" | sed 's/\.\///g' >> .pkgignore

$(LOCAL_GOPATH)/src:
	mkdir -p $(LOCAL_GOPATH)/src

$(LOCAL_GOPATH)/src/github.com/kr/pretty: $(LOCAL_GOPATH)/src
	GOPATH=$(LOCAL_GOPATH) go get github.com/kr/pretty

$(LOCAL_GOPATH)/src/code.google.com/p/goprotobuf/proto: $(LOCAL_GOPATH)/src
	GOPATH=$(LOCAL_GOPATH) go get code.google.com/p/goprotobuf/proto

$(LOCAL_GOPATH)/src/github.com/soundcloud/doozer: $(LOCAL_GOPATH)/src $(LOCAL_GOPATH)/src/github.com/kr/pretty $(LOCAL_GOPATH)/src/code.google.com/p/goprotobuf/proto
	GOPATH=$(LOCAL_GOPATH) go get github.com/soundcloud/doozer

local_build: update_version $(LOCAL_GOPATH)/src/github.com/soundcloud/doozer
	test -e bin || mkdir bin
	test -e $(VISOR_GO_PATH) || { mkdir -p $$(dirname $(VISOR_GO_PATH)); ln -sf $${PWD} $(VISOR_GO_PATH); }
	GOPATH=$(LOCAL_GOPATH) go build
	GOPATH=$(LOCAL_GOPATH) go build -o bin/visor ./cmd


########## packaging
FPM_EXECUTABLE:=$$(dirname $$(dirname $$(gem which fpm)))/bin/fpm
FPM_ARGS=-t deb -m 'Visor authors (see page), Daniel Bornkessel <daniel@soundcloud.com> (packaging)' --url http://github.com/soundcloud/visor -s dir
FAKEROOT=fakeroot
RELEASE=$$(cat .release 2>/dev/null || echo "0")

package: local_build
	- mkdir -p $(FAKEROOT)/usr/bin
	cp bin/visor $(FAKEROOT)/usr/bin
	-rm *.deb

	$(FPM_EXECUTABLE) -n "visor" \
		-C $(FAKEROOT) \
		--description "visor cli" \
		$(FPM_ARGS) -t deb -v $$(cat VERSION) --iteration $(RELEASE) .;


bump_package_release:
		echo $$(( $(RELEASE) + 1 )) > .release


