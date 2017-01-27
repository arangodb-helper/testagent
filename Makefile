PROJECT := testAgent
SCRIPTDIR := $(shell pwd)
ROOTDIR := $(shell cd $(SCRIPTDIR) && pwd)
VERSION:= $(shell cat $(ROOTDIR)/VERSION)
COMMIT := $(shell git rev-parse --short HEAD)

GOBUILDDIR := $(SCRIPTDIR)/.gobuild
SRCDIR := $(SCRIPTDIR)
BINDIR := $(ROOTDIR)/bin

ORGPATH := github.com/arangodb
ORGDIR := $(GOBUILDDIR)/src/$(ORGPATH)
REPONAME := $(PROJECT)
REPODIR := $(ORGDIR)/$(REPONAME)
REPOPATH := $(ORGPATH)/$(REPONAME)

GOPATH := $(GOBUILDDIR)
GOVERSION := 1.7.4-alpine

ifndef GOOS
	GOOS := linux
endif
ifndef GOARCH
	GOARCH := amd64
endif

BINNAME := testAgent-$(GOOS)-$(GOARCH)
BIN := $(BINDIR)/$(BINNAME)

SOURCES := $(shell find $(SRCDIR) -name '*.go')
TEMPLATES := $(shell find $(SRCDIR)/templates -name '*.tmpl')

.PHONY: all clean deps docker build build-local

all: build

clean:
	rm -Rf $(BIN) $(GOBUILDDIR)

local:
	@${MAKE} -B GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) build-local

build: $(BIN)

build-local: build 
	@ln -sf $(BIN) $(ROOTDIR)/testAgent

deps:
	@${MAKE} -B -s $(GOBUILDDIR)

$(GOBUILDDIR):
	@mkdir -p $(ORGDIR)
	@rm -f $(REPODIR) && ln -s ../../../.. $(REPODIR)
	GOPATH=$(GOBUILDDIR) go get -u github.com/jteeuwen/go-bindata/...
	GOPATH=$(GOBUILDDIR) go get github.com/fsouza/go-dockerclient
	GOPATH=$(GOBUILDDIR) go get github.com/juju/errgo
	GOPATH=$(GOBUILDDIR) go get github.com/op/go-logging
	GOPATH=$(GOBUILDDIR) go get github.com/spf13/cobra
	GOPATH=$(GOBUILDDIR) go get golang.org/x/sync/errgroup
	GOPATH=$(GOBUILDDIR) go get github.com/cenkalti/backoff
	GOPATH=$(GOBUILDDIR) go get gopkg.in/macaron.v1
	GOPATH=$(GOBUILDDIR) go get github.com/go-macaron/bindata

templates/templates.go: $(GOBUILDDIR) $(TEMPLATES)
	$(GOBUILDDIR)/bin/go-bindata -pkg templates -prefix templates -modtime 0 -o templates/templates.go templates/...

$(BIN): $(GOBUILDDIR) $(SOURCES) templates/templates.go
	@mkdir -p $(BINDIR)
	docker run \
		--rm \
		-v $(SRCDIR):/usr/code \
		-e GOPATH=/usr/code/.gobuild \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		-e CGO_ENABLED=0 \
		-w /usr/code/ \
		golang:$(GOVERSION) \
		go build -a -installsuffix netgo -tags netgo -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" -o /usr/code/bin/$(BINNAME) $(REPOPATH)

docker: build
	docker build -t arangodb/testAgent .

