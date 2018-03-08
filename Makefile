default: all

dep:
	@cd cmd/ && go get -d

all:
	go build cmd/spider.go
	go build cmd/test.go
	go build cmd/search.go

clean:
	rm -f spider test
