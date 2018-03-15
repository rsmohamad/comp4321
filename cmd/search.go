package main

import (
	"bufio"
	"comp4321/retrieval"
	"fmt"
	"os"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter search term: ")
	query, _ := reader.ReadString('\n')

	results := retrieval.Search(query)

	fmt.Println(results)
}
