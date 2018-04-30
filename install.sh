#!/bin/sh

mkdir -p ~/go/src/
mv comp4321/ ~/go/src
mv index.db ~/go/src/comp4321/
echo "Resolving dependencies"
cd ~/go/src/comp4321 && make dep