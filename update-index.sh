#!/bin/bash

helm repo index charts/ --url https://dtsulik.github.io/homelab-monorepo/ --merge index.yaml
