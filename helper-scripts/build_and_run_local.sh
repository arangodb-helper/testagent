#!/bin/bash
make docker && IMAGE=arangodb/testagent:latest ./run_local.sh