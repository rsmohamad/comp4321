default: all

dep:
	go get cmd/

all:
	go build cmd/spider.go
	go build cmd/test.go

clean:
	rm -f spider test
