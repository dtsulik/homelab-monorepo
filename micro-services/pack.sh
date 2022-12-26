#!/bin/bash

for d in */ ; do
    echo "Processing $d"
    cd $d
    helm dep up
    helm package . -d ../../charts
    cd ..
done