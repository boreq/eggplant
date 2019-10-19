#!/bin/bash
set -e

# Build frontend
echo "Running yarn build"
cd ../eggplant-frontend
rm -rf dist
yarn build

# Build backend
cd ../eggplant
echo "Running https://github.com/rakyll/statik"
statik -f -src=../eggplant-frontend/dist
