## Building instruction - tested on lab2 machines

Enter the following commands to build:

./install.sh
cd ~/go/src/comp4321
make

## Running the program

./spider
This will run the spider and replace the content of index.db with
30 pages from www.cse.ust.hk

./test
This will run the test program and produce spider_result.txt

## Inside the install.sh

The install.sh will unpack the .tar file to ~/go/src/
The source code is placed there so we can use `go dep` to resolve the dependencies.
