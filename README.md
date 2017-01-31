# Test agent 

The ArangoDB test agent is intended to run long duration tests on ArangoDB clusters.
During the test various 'user-like' operations are run, while the test-agent is 
introducing choas.

When a failure in one of the test operations is detected, the test-agent will log the 
failure, accompanied with all relevant information (such as database server log files).

## Chaos 

The test-agent will introduce the following kinds of chaos.

- [x] Restart a server, one of each type at a time 
- [x] Kill a server, one of each type at a time 
- [ ] DBServer is completely lost & will not return 
- [ ] Coordinator is completely lost & will not return 
- [x] Entire machine (with agent, dbserver & coordinator) is restarted 
- [ ] Entire machine (with dbserver & coordinator) is lost and replaced by another one 
- [ ] Entire machine (with dbserver & coordinator) is added 
- [ ] Entire machine (with dbserver & coordinator) is removed
- [ ] Network traffic between servers is blocked (iptables DENY)
- [ ] Network traffic between servers is ignored (iptables DROP)

It should also be possible to:

- [ ] Pause introducing chaos 
- [ ] Resume introducing chaos 

## Test operations 

The test agent will allow for multiple test scripts to be developed & run.
The test operations covered in those scripts will include (among others):

- [ ] Import documents 
- [x] Create documents
- [x] Read existing documents 
- [x] Read non-existing documents 
- [x] Remove existing documents 
- [x] Remove non-existing documents 
- [ ] Update existing documents 
- [ ] Update non-existing documents 
- [ ] Query documents (AQL)
- [ ] Backup entire databases

## Usage 

```
make docker
export IP=<your-local-IP>
docker run -it --rm -p 4200:4200 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    arangodb/testagent --docker-host-ip=$IP
```

Then connect your browser to http://localhost:4200 to see the test dashboard.

To run 'machines' on multiple physical machine, you must provide the endpoints of docker daemons running on 
these machines. E.g.

```
export IP=<your-local-IP>
docker run -it --rm -p 4200:4200 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    arangodb/testagent --docker-host-ip=$IP \
    --docker-endpoint=tcp://192.168.1.1:2376 \
    --docker-endpoint=tcp://192.168.1.2:2376 \
    --docker-endpoint=tcp://192.168.1.3:2376
```

### Options 

- `--agency-size number` Set the size of the agency for the new cluster.
- `--port` Set the first port used by the test agent (first of a range of ports). 
- `--log-level` Adjust log level (debug|info|warning|error)
- `--arangodb-image` Docker image containing `arangodb`. The image must exists in the local docker host.
- `--arango-image` Docker image containing `arangod`.
- `--docker-endpoint` How to reach the docker host (this option can be specified multiple times to use multiple docker hosts).
- `--docker-host-ip` IP of docker host.
