#!/bin/sh

mkdir -p ~/go/src/
tar -C ~/go/src -xaf comp4321.tar
echo "Resolving dependencies"
cd ~/go/src/comp4321 && make dep