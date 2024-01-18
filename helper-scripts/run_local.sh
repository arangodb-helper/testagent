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
       -v /tmp/testagent-reports/:/reports \
       -v /tmp/testagent-metrics/:/metrics \
       $IMAGE \
       --log-level=debug \
       --collect-metrics \
       --docker-host-ip=10.0.0.1 \
       --arango-image="$ARANGODB_IMAGE" \
       --arangodb-image="arangodb/arangodb-starter:$STARTER_VERSION" \
       --replication-version-2 \
       --replication2-operation-timeout 3m \
       --replication2-retry-timeout 8m \
       --replication2-document-size 256 \
       --replication2-batch-size 50 \
       --replication2-max-documents 10000 \
       --replication2-max-updates 3 \
       --graph-max-vertices 1000 \
       --graph-vertex-size 4096 \
       --graph-edge-size 4096 \
       --graph-traversal-ops 100 \
       --graph-batch-size 50
