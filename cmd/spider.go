package main

import (
	"comp4321/database"
	"comp4321/webcrawler"
	"fmt"
	"os"
	"time"
	"strconv"
)

func main() {
	index, _ := database.LoadIndexer("index.db")
	defer index.Close()
	index.DropAll()

	start := "http://www.cse.ust.hk"
	numPages := 30

	if len(os.Args) == 3{
		start = os.Args[1]
		numPages, _ = strconv.Atoi(os.Args[2])
	}

	startCrawl := time.Now()
	webcrawler.Crawl(start, numPages, index)
	elapsed := time.Since(startCrawl)
	fmt.Printf("Indexing %d pages took %s\n", numPages, elapsed)
}
