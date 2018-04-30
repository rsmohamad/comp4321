default: all

dep:
	@cd cmd/ && go get -d -v

all: dep
	go build cmd/spider.go
	go build cmd/server.go

tools:
	go build cmd/test.go
	go build cmd/print.go
	go build cmd/search.go

tests:
	go test ./database/ ./models/ ./retrieval/ ./stopword/ -cover

clean:
	rm -f spider test server search phase1.zip print

report:
	$(MAKE) -C reports

zip: clean report
	rm -rf phase2.zip
	git archive --format=zip --prefix=comp4321/ --output=phase2.zip HEAD
	zip phase2.zip readme.txt install.sh reports/phase2.pdf
	zip phase2.zip index.db
