#!/bin/bash
set -e

# Build frontend
echo "Running yarn build"
cd ../eggplant-frontend
rm -rf dist
yarn build

# Build backend
cd ../eggplant/ports/http/frontend
echo "Running https://github.com/rakyll/statik"
statik -f -src=../../../../eggplant-frontend/dist
