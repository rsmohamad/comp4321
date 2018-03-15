package webcrawler

import (
	"sync"
	"comp4321/database"
	"comp4321/models"
	"fmt"
)

func concurrentFetch(url string, results *chan *models.Document, wg *sync.WaitGroup) {
	page := Fetch(url)
	if page != nil {
		// skip if the page has no title and text
		if len(page.Words) == 0 && page.Title == "" {
			return
		}
		*results <- page
		wg.Done()
	}
}

func Crawl(uri string, num int, index *database.Indexer) []*models.Document {
	pages := make([]*models.Document, 0)
	visited := make(map[string]bool)
	results := make(chan *models.Document)
	fetchWg := sync.WaitGroup{}
	updateWg := sync.WaitGroup{}
	fetchWg.Add(num)

	// Visit first page
	go concurrentFetch(uri, &results, &fetchWg)
	visited[uri] = true

	for len(pages) < num {
		// append page from results channel
		page := <-results
		pages = append(pages, page)

		fmt.Printf("Fetched page #%d out of %d : %s\n", len(pages), num, page.Uri)

		updateWg.Add(1)
		go func(i int){
			index.UpdateOrAddPage(page)
			fmt.Printf("Indexed page #%d out of %d : %s\n", i, num, page.Uri)
			updateWg.Done()
		}(len(pages))

		// fetch all unvisited links
		for _, link := range page.Links {
			// skip if the link is already visited or indexed
			if visited[link] || index.ContainsUrl(link) {
				continue
			} else {
				go concurrentFetch(link, &results, &fetchWg)
				visited[link] = true
			}
		}
	}

	fetchWg.Wait()
	updateWg.Wait()
	fmt.Println("Done")
	return pages
}
