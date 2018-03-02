# COMP 4321 Search Engine

## Prerequisites

- Golang. The toolchain can be obtained from Go's website. If you are on Ubuntu and prefer using the package manager, you can run the following:
    ```
    sudo add-apt-repository ppa:longsleep/golang-backports
    sudo apt-get update
    sudo apt-get install golang-go
    ```

- After installing Go, set the GOPATH env variable to point to your go workspace. The default workspace location is `$HOME/go`. You can append `export GOPATH=$HOME/go` to your `.bashrc` file so that the GOPATH variable is always set. Additionally, you can use `go env` to check if your environment variables are set correctly.

- To use go tools for building and resolving dependencies, your application files must reside under `$GOPATH/src`. This repository must be cloned inside that directory.
    ```
    git clone https://github.com/rsmohamad/comp4321.git $GOPATH/src/comp4321
    ```


## Building

- Resolving dependencies
    ```
    make dep
    ```

- Build executables
    ```
    make
    ```

- Clean build files
    ```
    make clean
    ```

