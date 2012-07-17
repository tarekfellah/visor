# Makefile to include for Go projects
#
#   Installs dependencies in local GOPATH (.gopath),
#   and binary in $PWD/bin.
#
#   If you have sub folders with other packages or a
#   binary, set the env var SUB_PKG (for multiple separated
#   by blanks)
#
# usage:
#
#   PKG=github.com/org/project-name
#   SUB_PKG=github.com/org/project-name/cmd github.com/org/project-name/test-cmd
#
#   include go.mk
#
#   build: goclean gobuild
#   clean: goclean
#     rm -rf bin
#
LOCAL_GOPATH=${PWD}/.gopath
PKG_GOPATH=$(LOCAL_GOPATH)/src/$(PKG)

# this is needed for bazooka builds
unexport GIT_DIR


gobuild: gofmt
	mkdir -p $$(dirname $(PKG_GOPATH)); \
	ln -sfn $${PWD} $(PKG_GOPATH); \
	cd $(PKG_GOPATH);\
	  mkdir -p bin; \
	  GOPATH=$(LOCAL_GOPATH) go get -v -d;\
	  GOPATH=$(LOCAL_GOPATH) GOBIN=$${PWD}/bin go install -v $(PKG); \
	  test -z "$(SUB_PKG)" || GOPATH=$(LOCAL_GOPATH) GOBIN=$${PWD}/bin go install -v $(SUB_PKG); \

gofmt:
	GOPATH=$(LOCAL_GOPATH) go fmt ./...

gocheck:
	@test -n "$(PKG)" || { echo "PKG variable must be set to package name" && exit 1; }

goclean:
	rm -rf $(LOCAL_GOPATH)
