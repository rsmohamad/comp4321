#!/bin/sh

rm -f index.db

./spider -start=https://www.nytimes.com/ -pages=300 -a
./spider -start=https://www.theguardian.com/ -pages=300 -a
./spider -start=https://www.bbc.co.uk/ -pages=300 -a
./spider -start=https://www.cnn.com/ -pages=300 -a
