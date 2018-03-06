package webcrawler

import (
	"sync"
	"comp4321/database"
	"comp4321/models"
)

func concurrentUpdate(page *models.Document, index *database.Indexer, wg *sync.WaitGroup) {
	index.UpdateOrAddPage(page)
	wg.Done()
}

func concurrentFetch(url string, results *chan *models.Document, wg *sync.WaitGroup) {
	page := Fetch(url)
	if page != nil {
		// skip if the page has no title and text
		if len(page.Words) == 0 && page.Title == ""{
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
	wg := sync.WaitGroup{}
	wg.Add(num)

	// Visit first page
	go concurrentFetch(uri, &results, &wg)
	visited[uri] = true

	for len(pages) < num {
		// append page from results channel
		page := <-results
		pages = append(pages, page)

		// update page to index
		wg.Add(1)
		go concurrentUpdate(page, index, &wg)

		// fetch all unvisited links
		for _, link := range page.Links {
			// skip if the link is already visited or indexed
			if visited[link] || index.IsUrlPresent(link) {
				continue
			} else {
				go concurrentFetch(link, &results, &wg)
				visited[link] = true
			}
		}
	}

	wg.Wait()
	return pages
}
