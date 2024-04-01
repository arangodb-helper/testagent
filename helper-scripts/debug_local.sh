#!/bin/bash
make docker-dbg && IMAGE=arangodb/testagent:dbg ./helper-scripts/run_local.sh