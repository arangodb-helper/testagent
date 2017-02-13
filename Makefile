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

ifndef DOCKERNAMESPACE
	DOCKERNAMESPACE := arangodb
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
	@rm -f $(GOBUILDDIR)/src/github.com/jteeuwen && ln -s ../../../vendor/github.com/jteeuwen $(GOBUILDDIR)/src/github.com/jteeuwen

templates/templates.go: $(GOBUILDDIR) $(TEMPLATES)
	GOPATH=$(GOBUILDDIR) go build -o $(GOBUILDDIR)/bin/go-bindata github.com/jteeuwen/go-bindata/go-bindata
	$(GOBUILDDIR)/bin/go-bindata -pkg templates -prefix templates -modtime 1486974991 -ignore templates.go -o templates/templates.go templates/...

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
	docker build -t arangodb/testagent .

docker-push: docker
ifneq ($(DOCKERNAMESPACE), arangodb)
	docker tag arangodb/testagent $(DOCKERNAMESPACE)/testagent
endif
	docker push $(DOCKERNAMESPACE)/testagent

localtest:
	docker run -it --rm --net=host -v $(HOME)/tmp:/reports -v /var/run/docker.sock:/var/run/docker.sock arangodb/testagent --docker-net-host

docker-push-version: docker
	docker tag arangodb/testagent arangodb/testagent:$(VERSION)
	docker push arangodb/testagent:$(VERSION)

release-patch: $(GOBUILDDIR)
	go run ./tools/release.go -type=patch 

release-minor: $(GOBUILDDIR)
	go run ./tools/release.go -type=minor

release-major: $(GOBUILDDIR)
	go run ./tools/release.go -type=major 
