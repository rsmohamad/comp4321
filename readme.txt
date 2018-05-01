--- Building (tested on lab2 machines) ---

./install.sh
cd ~/go/src/comp4321
make

--- Running the program ---

./spider
./server

--- Notes ---

- The project requires Go compiler. Lab2 has version 1.9.
- The source files must be inside $GOPATH/src
  for the make script to work correctly/.
- Lab2 machines $GOPATH points to ~/go

--- Installing Go ---

In case the system does not have Go installed:

wget https://dl.google.com/go/go1.10.linux-amd64.tar.gz
sudo tar -C /usr/local/ -xzf go1.10.linux-amd64.tar.gz
echo "export PATH=\$PATH:/usr/local/go/bin" | sudo tee -a /etc/profile
source /etc/profile

Sorry for the extra procedures.
