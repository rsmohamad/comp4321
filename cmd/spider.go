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

	startCrawl := time.Now()
	webcrawler.Crawl("http://www.cse.ust.hk", 30, index)
	elapsed := time.Since(startCrawl)
	fmt.Printf("Indexing %d pages took %s\n", 30, elapsed)
}
