PROJECT := testAgent
SCRIPTDIR := $(shell pwd)
ROOTDIR := $(shell cd $(SCRIPTDIR) && pwd)
VERSION:= $(shell cat $(ROOTDIR)/VERSION)
COMMIT := $(shell git rev-parse --short HEAD)

SRCDIR := $(SCRIPTDIR)
BINDIR := $(ROOTDIR)/bin

ORGPATH := github.com/arangodb-helper
REPONAME := $(PROJECT)
REPOPATH := $(ORGPATH)/$(REPONAME)

GOVERSION := 1.22.1-alpine

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

.PHONY: all clean deps docker build build-local tests

all: docker

clean:
	rm -Rf $(BIN)

templates/templates.go: $(TEMPLATES)
	go build -o bin/go-bindata github.com/jteeuwen/go-bindata/go-bindata
	bin/go-bindata -pkg templates -prefix templates -modtime 1486974991 -ignore templates.go -o templates/templates.go templates/...

docker: $(SOURCES) templates/templates.go
	docker build \
		--build-arg GOVERSION=$(GOVERSION) \
		--build-arg GOOS="$(GOOS)" \
		--build-arg GOARCH="$(GOARCH)" \
		--build-arg BINNAME="$(BINNAME)" \
		--build-arg GOTAGS="$(GOTAGS)" \
		--build-arg VERSION="$(VERSION)" \
		--build-arg COMMIT="$(COMMIT)" \
		-f Dockerfile -t arangodb/testagent .

docker-dbg: $(SOURCES) templates/templates.go
	docker build \
		--build-arg GOVERSION=$(GOVERSION) \
		--build-arg GOOS="$(GOOS)" \
		--build-arg GOARCH="$(GOARCH)" \
		--build-arg BINNAME="$(BINNAME)" \
		--build-arg GOTAGS="$(GOTAGS)" \
		--build-arg VERSION="$(VERSION)" \
		--build-arg COMMIT="$(COMMIT)" \
		-f Dockerfile.debug -t arangodb/testagent:dbg .

docker-push: docker
ifneq ($(DOCKERNAMESPACE), arangodb)
	docker tag arangodb/testagent $(DOCKERNAMESPACE)/testagent
endif
	docker push $(DOCKERNAMESPACE)/testagent

localtest:
	docker run -it --rm --net=host -v $(HOME)/tmp:/reports -v /var/run/docker.sock:/var/run/docker.sock arangodb/testagent --docker-net-host

tests:
	go test -coverprofile cover.out github.com/arangodb-helper/testagent/tests/simple -v
	go tool cover -html=cover.out

docker-push-version: docker
	docker tag arangodb/testagent arangodb/testagent:$(VERSION)
	docker push arangodb/testagent:$(VERSION)

release-patch:
	go run ./tools/release.go -type=patch 

release-minor:
	go run ./tools/release.go -type=minor

release-major:
	go run ./tools/release.go -type=major 
