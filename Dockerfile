ARG BINNAME
ARG GOVERSION=1.20-alpine
FROM golang:${GOVERSION} AS downloader

# git is required by 'go mod'
RUN apk add git

WORKDIR $GOPATH/src/github.com/arangodb-helper/testagent

COPY go.mod .
COPY go.sum .
# It is done only once unless go.mod has been changed
RUN go mod download


FROM downloader AS builder
ARG VERSION
ARG COMMIT
ARG BINNAME
ARG GOARCH
ARG GOOS
ARG GOTAGS

COPY *.go ./
COPY pkg pkg
COPY service service
COPY templates templates
COPY tests tests

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOARCH=${GOARCH}
ENV GOOS=${GOOS}

RUN go build -installsuffix netgo -tags "${GOTAGS}" \
    -ldflags "-extldflags '-static' -X main.projectVersion=${VERSION} -X main.projectBuild=${COMMIT}" -o /${BINNAME}

FROM alpine:3.18.3
ARG BINNAME
COPY --from=builder /${BINNAME} /app/testAgent
EXPOSE 4200

# Signal running in docker 
ENV RUNNING_IN_DOCKER=true

# Image containing arangodb starter 
ENV ARANGODB_IMAGE=arangodb/arangodb-starter:latest

# Database image 
#ENV ARANGO_IMAGE=arangodb/arangodb:3.1.19
#ENV ARANGO_IMAGE=arangodb/arangodb-preview:3.4.0-rc.3
#ENV ARANGO_IMAGE=neunhoef/arangodb-community:3.4.0-rc4
ENV ARANGO_IMAGE=arangodb/enterprise:3.9.1

# network-blocker image
ENV NETWORK_BLOCKER_IMAGE=arangodb/network-blocker:0.1.0

# Failure reports dir 
ENV REPORT_DIR=/reports
VOLUME /reports

ENTRYPOINT ["/app/testAgent"]
