#!/bin/sh

rm -f index.db

./spider https://www.nytimes.com/ 300
./spider https://www.theguardian.com/ 300
./spider https://www.bbc.co.uk/ 300
./spider https://www.cnn.com/ 300
