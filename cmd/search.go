package main

import (
	"comp4321/retrieval"
	"fmt"
)

func main() {
	query := "computer business"
	results := retrieval.Search(query)

	fmt.Println(results)
}
