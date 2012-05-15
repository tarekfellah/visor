LOCAL_GOPATH=${PWD}/local_go_path

$(LOCAL_GOPATH)/src:
	mkdir -p $(LOCAL_GOPATH)/src

$(LOCAL_GOPATH)/src/github.com/soundcloud/doozer: $(LOCAL_GOPATH)/src
	GOPATH=$(LOCAL_GOPATH) go get github.com/soundcloud/doozer

$(LOCAL_GOPATH)/src/github.com/kesselborn/go-getopt: $(LOCAL_GOPATH)/src
	GOPATH=$(LOCAL_GOPATH) go get github.com/kesselborn/go-getopt

build: fmt $(LOCAL_GOPATH)/src/github.com/soundcloud/doozer $(LOCAL_GOPATH)/src/github.com/kesselborn/go-getopt
	GOPATH=$(LOCAL_GOPATH) go install
	GOPATH=$(LOCAL_GOPATH) go build -o bin/visor ./cmd

compile: fmt
	go build
	go build -o bin/visor ./cmd

fmt:
	go fmt ./...

clean:
	$(LOCAL_GOPATH)/src
