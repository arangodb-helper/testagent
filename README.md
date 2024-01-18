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
- [x] Entire machine (with agent, dbserver & coordinator) is restarted 
- [ ] Entire machine (with agent, dbserver & coordinator) is replaced (currently impossible)
- [x] Entire machine (with dbserver & coordinator) is lost and replaced by another one 
- [x] Entire machine (with dbserver & coordinator) is added 
- [x] Entire machine (with dbserver & coordinator) is removed
- [x] Network traffic between servers is blocked (iptables REJECT)
- [x] Network traffic between servers is ignored (iptables DROP)
- [ ] Split brain

It should also be possible to:

- [x] Pause introducing chaos 
- [x] Resume introducing chaos 

## Test operations 

The test agent will allow for multiple test scripts to be developed & run.
The test operations covered in those scripts will include (among others):

- [x] Create collections 
- [x] Drop collections
- [x] Import documents 
- [x] Create documents
- [x] Read existing documents 
- [x] Read non-existing documents 
- [x] Remove existing documents 
- [x] Remove existing documents with explicit & last revision
- [x] Remove existing documents with explicit & non-last revision
- [x] Remove non-existing documents 
- [x] Update existing documents 
- [x] Update existing documents with explicit & last revision
- [x] Update existing documents with explicit & non-last revision
- [x] Update non-existing documents 
- [x] Replace existing documents 
- [x] Replace existing documents with explicit & last revision
- [x] Replace existing documents with explicit & non-last revision
- [x] Replace non-existing documents 
- [x] Query documents (AQL)
- [x] Query documents with long running query (AQL SLEEP)
- [x] Modify documents with query (AQL)
- [x] Modify documents with long running query (AQL SLEEP)
- [ ] Backup entire databases (export is not yet available on clusters)
- [x] Rebalance shards

## Usage 

```
make docker
export IP=<your-local-IP>
docker run -it --rm -p 4200:4200 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    arangodb/testagent --docker-host-ip=$IP
```

Then connect your browser to http://localhost:4200 to see the test dashboard.

To run 'machines' on multiple physical machine, you must provide the endpoints
of docker daemons running on these machines. E.g.

```
export IP=<your-local-IP>
docker run -it --rm -p 4200:4200 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    arangodb/testagent --docker-host-ip=$IP \
    --docker-endpoint=tcp://192.168.1.1:2376 \
    --docker-endpoint=tcp://192.168.1.2:2376 \
    --docker-endpoint=tcp://192.168.1.3:2376
```

To allow for remote access to the remote docker agents you might need to
add the `-H tcp://0.0.0.0:2376 --storage-driver=overlay2` to the `ExecStart`
line in your docker.service file.

To allow for use of TLS verified Docker (requires ArangoDB starter
version > 0.13.10) service export additionally the relevant default
Docker environment variable for enabling the verification before the
above Docker command..

```
export DOCKER_TLS_VERIFIED=1
```

The above assumes that the relevant `ca.cert`, `cert.pem` and
`key.pem` reside in the default location for Docker client
certification, `$HOME/.docker`. If you would like to store the
certificate in a different directory, it needs to be specified
accordingly:

```
export DOCKER_CERT_PATH=/path/to/cert
```

### Options 

- `--agency-size number` Set the size of the agency for the new cluster.
- `--port` Set the first port used by the test agent (first of a range of ports). 
- `--log-level` Adjust log level (debug|info|warning|error)
- `--chaos-level` Chaos level. Allowed values: 0-4. 0 = no chaos. 4 = maximum chaos. Default: 4.
- `--arangodb-image` Docker image containing `arangodb`. The image must exists in the local docker host.
- `--arango-image` Docker image containing `arangod`.
- `--docker-endpoint` How to reach the docker host (this option can be specified multiple times to use multiple docker hosts).
- `--docker-host-ip` IP of docker host.
- `--docker-net-host` If set, run all containers with `--net=host`. (Make sure the testagent container itself is also started with `--net=host`). Network chaos is not supported with host networking.
- `--replication-version-2` If set, use replication version 2
