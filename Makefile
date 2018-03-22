default: all

dep:
	@cd cmd/ && go get -d

all:
	go build cmd/spider.go
	go build cmd/test.go
	go build cmd/search.go
	go build cmd/server.go
	go build cmd/gojek.go

clean:
	rm -f spider test server
