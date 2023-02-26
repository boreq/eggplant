#!/bin/bash
set -e

versionFile="ports/http/frontend/version.go"
previousCommit=$(cat ${versionFile} | grep FrontendCommit | awk ' {gsub(/"/, "", $4); print $4 }')

cd ../eggplant-frontend

if ! git diff-index --quiet HEAD --; then
    echo "You have uncommited changes in the frontend repository!"
    exit 1
fi

commit=$(git rev-parse HEAD)
echo "Previous frontend commit: ${previousCommit}"
echo "Current frontend commit: ${commit}"

if [ "$commit" = "$previousCommit" ]; then
    echo "Frontend is already up to date (${commit} == ${previousCommit})"
    exit 1
fi

echo "Generating a commit message ${previousCommit} -> ${commit}"

commitFile="/tmp/eggplant-frontend-commit.txt"
touch ${commitFile}
echo "Update frontend" > ${commitFile}
echo "" >> ${commitFile}
git log --pretty=medium ${previousCommit}..${commit} >> ${commitFile}

echo "Building the frontend"
rm -rf dist
NODE_OPTIONS=--openssl-legacy-provider yarn build

cd ../eggplant

if ! git diff-index --quiet HEAD --; then
    echo "You have uncommited changes in the backend repository!"
    exit 1
fi

echo "Copying build files"
cd ./ports/http/frontend
rm -r ./img ./css ./js
cp -r ../../../../eggplant-frontend/dist/. ./
cd ../../../

echo "Persisting frontend version"

echo "package frontend" > "${versionFile}"
echo "" >> "${versionFile}"
echo "const FrontendCommit = \"${commit}\"" >> "${versionFile}"

git add ports/http/frontend/
git commit -F ${commitFile}
