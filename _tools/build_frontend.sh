#!/bin/bash
set -e

# Build frontend
echo "Running yarn build"
cd ../plum-frontend
rm -rf dist
yarn build

# Build backend
cd ../plum
echo "Running https://github.com/rakyll/statik"
statik -f -src=../plum-frontend/dist
