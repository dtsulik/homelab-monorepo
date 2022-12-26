#!/bin/bash

helm repo index charts/ --url https://dtsulik.github.io/homelab-monorepo/charts --merge index.yaml
mv charts/index.yaml .
