MAJOR    := 0
MINOR    := 8
PATCH    := 0
VERSION  := $(MAJOR).$(MINOR).$(PATCH)
LDFLAGS  := -ldflags "-X main.VERSION $(VERSION)"

default:
	go build $(LDFLAGS)

