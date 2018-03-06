# COMP 4321 Search Engine

## Getting started

- Install golang. If you are on Ubuntu and prefer using the package manager, you can run the following:
    ```
    sudo add-apt-repository ppa:longsleep/golang-backports
    sudo apt-get update
    sudo apt-get install golang-go
    ```
    
 - If you are on the CentOS VM, you can download the binaries from go website:
   ```
   wget https://dl.google.com/go/go1.10.linux-amd64.tar.gz
   tar -C /usr/local/ -xzf go1.10.linux-amd64.tar.gz
   ```
   Append `export PATH=$PATH:/usr/local/go/bin` into `/etc/profile` and then run `source /etc/profile`.
   
- Make sure the `GOPATH` environment variable is set. You can check by running `go env`. The default value should be `/home/<username>/go`.

- To use go tools for building and resolving dependencies, your project must reside under `$GOPATH/src`. This repository must be cloned inside that directory.
    ```
    git clone https://github.com/rsmohamad/comp4321.git $GOPATH/src/comp4321
    ```


## Building

- Inside the project directory, type
    ```
    make dep
    make
    ```

