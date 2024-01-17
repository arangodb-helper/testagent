#!/bin/bash
make docker-dbg && IMAGE=arangodb/testagent:dbg ./run_local.sh