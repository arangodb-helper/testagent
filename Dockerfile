FROM alpine:3.4

ADD ./bin/testAgent-linux-amd64 /app/testAgent

EXPOSE 4200

# Signal running in docker 
ENV RUNNING_IN_DOCKER=true

# Image containing arangodb starter 
ENV ARANGODB_IMAGE=arangodb/arangodb-starter:0.3.1

# Database image 
#ENV ARANGO_IMAGE=arangodb/arangodb:3.1.10
ENV ARANGO_IMAGE=neunhoef/arangodb:3.2.devel

# network-blocker image
ENV NETWORK_BLOCKER_IMAGE=arangodb/network-blocker:0.0.1

# Failure reports dir 
ENV REPORT_DIR=/reports
VOLUME /reports

ENTRYPOINT ["/app/testAgent"]