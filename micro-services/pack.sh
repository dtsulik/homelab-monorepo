#!/bin/bash

set -e -o pipefail

for d in */ ; do
    echo "Processing $d"
    cd $d
    helm dep up
    helm package . -d ../../charts
    cd ..
done