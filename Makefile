default: all

dep:
	@cd cmd/ && go get -d

all: dep
	go build cmd/spider.go
	go build cmd/test.go
	go build cmd/search.go
	go build cmd/server.go

clean:
	rm -f spider test server search phase1.zip

report:
	$(MAKE) -C reports

zip: clean report
	rm -rf comp4321/ phase1/
	mkdir -p comp4321 phase1
	cp -r cmd/ database/ models/ retrieval/ static/ stopword/ views/ webcrawler/ comp4321/
	cp Makefile README.md index.db spider_result.txt comp4321/
	tar -caf comp4321.tar comp4321/
	rm -rf comp4321/
	mv comp4321.tar phase1/
	cp readme.txt install.sh reports/*.pdf phase1/
	zip -r phase1.zip phase1/
	rm -rf phase1/
