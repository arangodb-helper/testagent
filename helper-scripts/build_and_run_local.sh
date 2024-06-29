#!/bin/bash
make docker && IMAGE=arangodb/testagent:latest ./helper-scripts/run_local.sh