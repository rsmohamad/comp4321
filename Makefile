default: all

dep:
	@cd cmd/ && go get -d

all: dep
	go build cmd/spider.go
	go build cmd/test.go
	go build cmd/search.go
	go build cmd/server.go

clean:
	rm -f spider test server search

report:
	$(MAKE) -C reports

zip: clean report
	rm -rf comp4321/
	mkdir comp4321
	cp -r cmd/ database/ models/ retrieval/ static/ stopword/ views/ webcrawler/ comp4321/
	cp Makefile README.md index.db spider_result.txt comp4321/
	cp reports/*.pdf comp4321/
	tar -caf comp4321.tar comp4321/
	rm -rf comp4321/
	zip comp4321.zip readme.txt install.sh comp4321.tar
	rm comp4321.tar
