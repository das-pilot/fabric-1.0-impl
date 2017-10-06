#!/bin/bash -eu
if [ ! -d "fabric" ]; then
    git clone https://github.com/hyperledger/fabric.git
fi
docker-compose up
docker rm listener-build