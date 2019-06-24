#!/bin/bash

cache_file=$(find . -name "*~")
cache_file+=' '
cache_file+=$(find . -name "*.sv*")
cache_file+=' '
cache_file+=$(find . -name "*.swp")
cache_file+=' '
cache_file+=$(find . -name "*.sw*")
cache_file+=' '
cache_file+=$(find . -name "*.py[co]")
cache_file+=' '
cache_file+=$(find . -name "*.bin")
cache_file+=' '
cache_file+=$(find . -name "*.log")
cache_file+=' '
cache_file+=$(find . -name "*.png")
cache_file+=' tmp/* !(tmp/.gitignore)'

for cache in ${cache_file}
do
    echo "Deleting file ${cache}" && rm -rf ${cache}
done

