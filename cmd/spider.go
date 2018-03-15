package main

import (
	"fmt"
	"time"
	"comp4321/database"
	"comp4321/webcrawler"
)

func main() {
	index, _ := database.LoadIndexer("index.db")
	defer index.Close()
	index.DropAll()

	start := time.Now()
	webcrawler.Crawl("https://www.nytimes.com/", 300, index)
	elapsed := time.Since(start)
	fmt.Printf("Indexing %d pages took %s\n", 300, elapsed)
}
