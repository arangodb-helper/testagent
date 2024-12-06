# Test agent 

The ArangoDB test agent is intended to run long duration tests on ArangoDB clusters.
During the test various 'user-like' operations are run, while the test-agent is 
introducing chaos.

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
- [x] Create graphs
- [x] Add vertex and edge documents to existing graphs
- [x] Traverse graphs

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

## Tests
### simple
This is the first test introduced, and the only one available in versions below 1.1.0. The test performs various operations on collections and documents within the _system databse, such as:
* Create collections
* Remove collections
* Create documents
* Import documents
* Read documents
* Update documents via API
* Update documents via AQL queries
* Replace documents
* Query documents  

Actions are executed in random order.


All the tests except `simple` make a suite that is named `complex`. These tests share common actions and settings. They are contained in the `complex` package.  
Unlike `simple`, `complex` test do not execute actions in random order. That is because those tests are intended to find bugs that might manifest only when a deployment contains certain amounts of data.  
### DocColTest and OneShardTest
These tests perform various operations on collections and documents within a databse:
* Create database
* Create collection
* Create documents
* Read documents
* Update documents
* Drop collection
* Drop database  

Each document contains a text field named `payload` that contains a random sequence of charactes. Its length is configurable and is limited only by disk space available. When documents are read, the random sequence is re-generated from the same seed and the result is compared to the one read from the database.  
The only difference between `DocColTest` and `OneShardTest` is that the latter uses a single-sharded database.  

### Graph tests
The are 3 tests that work with graphs, that differ only in the type of graph used: `CommunityGraphTest`, `SmartGraphTest` and `EnterpriseGraphTest`. Graph tests perform the following actions:
* Create graph and underlying collections
* Create vertices
* Create edges
* Traverse graph
* Drop graph and collections  
Both vertex and edge documents contain a `payload` field of configurable size.

## Options 

### General
- `--agency-size number` Set the size of the agency for the new cluster.
- `--port` Set the first port used by the test agent (first of a range of ports). 
- `--log-level` Adjust log level (debug|info|warning|error)
- `--chaos-level` Chaos level. Allowed values: 0-4. 0 = no chaos. 4 = maximum chaos. Default: 4.
- `--arangodb-image` Docker image containing `arangodb`. The image must exists in the local docker host.
- `--arango-image` Docker image containing `arangod`.
- `--docker-endpoint` How to reach the docker host (this option can be specified multiple times to use multiple docker hosts).
- `--docker-host-ip` IP of docker host.
- `--docker-net-host` If set, run all containers with `--net=host`. (Make sure the testagent container itself is also started with `--net=host`). Network chaos is not supported with host networking.
- `--force-one-shard` If set, force one shard arangodb cluster (default: false)
- `--replication-version-2` If set, use replication version 2
- `--return-403-on-failed-write-concern` If set, option `--cluster.failed-write-concern-status-code` will not be set for DB servers. Otherwise this parameter will be set to 503. Warning: if this option is set, getting a response 403 from coordinator will be treated as a failure. (default: false)
- `--docker-interface` Network interface used to connect docker containers to (default: docker0)
- `--report-dir` Directory in which failure reports will be created. This option can also be set with environment variable `REPORT_DIR`. The CLI parameter has higher priority than the envrironment variable. (default: .)
- `--collect-metrics` If set, metrics about docker containers will be collected and saved into files. List of metrics that are collected: `cpu_total_usage`, `cpu_usage_in_kernelmode`, `cpu_usage_in_usermode`, `system_cpu_usage`, `memory_usage`
- `--metrics-dir` Directory in which metrics will be stored.  This option can also be set with environment variable `METRICS_DIR`. The CLI parameter has higher priority than the envrironment variable. (default: .)
- `--privileged` If set, run all containers with `--privileged`
- `--max-machines` Upper limit to the number of machines in a cluster (default: 10)
- `enable-test` Enable particular test. This option can be specified multiple times to run multiple tests simultaneously. Default: run all tests. Available tests: `simple`, `DocColTest`, `OneShardTest`, `CommunityGraphTest`, `SmartGraphTest`, `EnterpriseGraphTest`


### Test-specific
Options starting with `--simple` affect only the simple test.  
All options starting with `--complex` affect all the tests in the `complex` suite.  
All options starting with `--doc` affect all the "document" tests in the `complex` suite(`DocColTest` and `OneShardTest`)  
All options starting with `--graph` affect all the "graph" tests in the `complex` suite(`CommunityGraphTest`, `SmartGraphTest`, `EnterpriseGraphTest`)  
- `--simple-max-documents` Upper limit to the number of documents created in simple test
- `--simple-max-collections` Upper limit to the number of collections created in simple test
- `--simple-operation-timeout` Timeout per database operation
- `--simple-retry-timeout` How long are tests retried before giving up
- `--complex-shards` Number of shards
- `--complex-replicationFactor` Replication factor
- `--complex-operation-timeout` Timeout per database operation
- `--complex-retry-timeout` How long are tests retried before giving up
- `--complex-step-timeout` Pause between test actions
- `--doc-max-documents` Upper limit to the number of documents created in document collection tests
- `--doc-batch-size` Number of documents to be created during one test step
- `--doc-document-size` Size of payload field in bytes
- `--doc-max-updates` Number of update operations to be performed on each document
- `--graph-max-vertices` Upper limit to the number of vertices
- `--graph-vertex-size` Size of the payload field in bytes in all vertices
- `--graph-edge-size` Size of the payload field in bytes in all edges
- `--graph-traversal-ops` How many traversal operations to perform in one test step
- `--graph-batch-size` Number of vertices/edges to be created in one test step
