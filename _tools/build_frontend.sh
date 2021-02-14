#!/bin/bash
set -e

versionFile="ports/http/frontend/version.go"
previousCommit=$(cat ${versionFile} | grep -Eo '[0-9a-f]{32}')


cd ../eggplant-frontend

if ! git diff-index --quiet HEAD --; then
    echo "You have uncommited changes in the frontend repository!"
    exit 1
fi

commit=$(git rev-parse HEAD)
echo "Current frontend commit: ${commit}"

echo "Generating a commit message ${previousCommit} -> ${commit}"

commitFile="/tmp/eggplant-frontend-commit.txt"
touch ${commitFile}
echo "Update frontend" > ${commitFile}
echo "" >> ${commitFile}
git log --pretty=medium ${previousCommit}..${commit} >> ${commitFile}

echo "Building the frontend"
rm -rf dist
yarn build

cd ../eggplant

if ! git diff-index --quiet HEAD --; then
    echo "You have uncommited changes in the backend repository!"
    exit 1
fi

echo "Running https://github.com/rakyll/statik"
cd ./ports/http/frontend
statik -f -src=../../../../eggplant-frontend/dist
cd ../../../

echo "Persisting frontend version"

echo "package frontend" > "${versionFile}"
echo "" >> "${versionFile}"
echo "const FrontendCommit = \"${commit}\"" >> "${versionFile}"

git add -u
git commit -F ${commitFile}
