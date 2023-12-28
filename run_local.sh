#!/bin/bash

if test -z "$STARTER_VERSION"; then
    STARTER_VERSION=0.17.2
fi

if test -z "$IMAGE"; then
    IMAGE=arangodb/testagent:latest
fi

if test -z "$ARANGODB_IMAGE"; then
    # ARANGODB_IMAGE=arangodb:3.11.6
    ARANGODB_IMAGE=arangodb/enterprise-preview:devel-nightly-amd64
fi

docker pull $ARANGODB_IMAGE

docker run -it --rm -p 4200:4200 -p 4000:4000 \
       --name testagent \
       -v /var/run/docker.sock:/var/run/docker.sock \
       -v /vitaly/projects/testagent/certdir:/certdir \
       -e DOCKER_CERT_PATH=/certdir \
       -e DOCKER_TLS_VERIFY=1 \
       $IMAGE \
       --log-level=debug \
       --docker-host-ip=10.0.0.1 \
       --arango-image="$ARANGODB_IMAGE" \
       --arangodb-image="arangodb/arangodb-starter:$STARTER_VERSION" \
       --replication2-document-size 32 \
       --replication2-batch-size 50 \
       --replication2-max-documents 100 \
       --graph-max-vertices 100 \
       --graph-vertex-size 512 \
       --graph-edge-size 512 \
       --graph-traversal-ops 100 \
       --graph-batch-size 50
