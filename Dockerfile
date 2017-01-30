FROM alpine:3.4

ADD ./bin/testAgent-linux-amd64 /app/testAgent

# Image containing arangodb starter 
#ENV ARANGODB_IMAGE=arangodb/arangodb-starter
ENV ARANGODB_IMAGE=ewoutp/arangodb-starter

# Database image 
#ENV ARANGO_IMAGE=arangodb/arangodb:3.1.9
ENV ARANGO_IMAGE=neunhoef/arangodb:3.1.devel

# Failure reports dir 
ENV REPORT_DIR=/reports
VOLUME /reports

ENTRYPOINT ["/app/testAgent"]