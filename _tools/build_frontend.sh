#!/bin/bash
set -e

# Build frontend
cd ../goaccess-frontend
rm -rf dist
yarn build

# Build backend
cd ../goaccess-backend
statik -src=../goaccess-frontend/dist
