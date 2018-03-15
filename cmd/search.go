package main

import (
	"bufio"
	"comp4321/retrieval"
	"fmt"
	"os"
)

func main() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter search term: ")
		query, _ := reader.ReadString('\n')

		results := retrieval.Search(query)

		for _, doc := range results {
			fmt.Println(doc.Title)
		}
	}
}
