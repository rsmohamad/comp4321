package main

import (
	"comp4321/database"
	"comp4321/webcrawler"
	"fmt"
	"time"
)

func main() {
	index, _ := database.LoadIndexer("index.db")
	defer index.Close()
	index.DropAll()

	numPages := 30
	start := "http://www.cse.ust.hk"

	startCrawl := time.Now()
	webcrawler.Crawl(start, numPages, index)
	elapsed := time.Since(startCrawl)
	fmt.Printf("Indexing %d pages took %s\n", numPages, elapsed)
}
