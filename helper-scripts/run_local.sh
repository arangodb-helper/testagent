#!/bin/bash

if test -z "$STARTER_VERSION"; then
    STARTER_VERSION=0.18.0
fi

if test -z "$IMAGE"; then
    IMAGE=arangodb/testagent:latest
fi

if test -z "$ARANGODB_IMAGE"; then
    ARANGODB_IMAGE=arangodb/enterprise:3.11.8
    # ARANGODB_IMAGE=arangodb/enterprise-test:devel-nightly-amd64
fi

docker pull $ARANGODB_IMAGE

docker run -it --rm -p 4200:4200 -p 4000:4000 \
       --name testagent \
       -v /var/run/docker.sock:/var/run/docker.sock \
       -v /tmp/testagent-certdir:/certdir \
       -e DOCKER_CERT_PATH=/certdir \
       -e DOCKER_TLS_VERIFY=1 \
       -v /tmp/testagent-reports/:/reports \
       -v /tmp/testagent-metrics/:/metrics \
       $IMAGE \
       --log-level=debug \
       --chaos-level=0 \
       --collect-metrics \
       --docker-host-ip=10.0.0.1 \
       --arango-image="$ARANGODB_IMAGE" \
       --arangodb-image="arangodb/arangodb-starter:$STARTER_VERSION" \
       --complex-step-timeout 5s \
       --complex-operation-timeout 3m \
       --complex-retry-timeout 8m \
       --doc-document-size 512 \
       --doc-batch-size 100 \
       --doc-max-documents 400 \
       --doc-max-updates 2 \
       --graph-max-vertices 1000 \
       --graph-vertex-size 4096 \
       --graph-edge-size 4096 \
       --graph-traversal-ops 100 \
       --graph-batch-size 50
