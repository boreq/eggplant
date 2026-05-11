#!/bin/bash

if [ -n "$(git status backend --porcelain 2>&1)" ]; then
    echo "Detected changes in the repository!"; 
    git --no-pager diff; 
    exit 1; 
else 
    echo "No changes detected in the repository."; 
fi
