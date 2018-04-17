package main

import (
	"comp4321/database"
	"comp4321/webcrawler"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	index, _ := database.LoadIndexer("index.db")
	defer index.Close()

	start := "https://www.cse.ust.hk/"
	numPages := 10000
	restrict := true

	if len(os.Args) == 3 {
		start = os.Args[1]
		numPages, _ = strconv.Atoi(os.Args[2])
		restrict = false
	}

	startCrawl := time.Now()
	obtained := webcrawler.Crawl(start, numPages, index, restrict)
	elapsed := time.Since(startCrawl)
	fmt.Printf("Indexing %d pages took %s\n", len(obtained), elapsed)
	fmt.Println("Updating term weights...")
	index.UpdateTermWeights()
	fmt.Println("Updating adj list...")
	index.UpdateAdjList()
	fmt.Println("Updating page rank...")
	index.UpdatePageRank()
}
