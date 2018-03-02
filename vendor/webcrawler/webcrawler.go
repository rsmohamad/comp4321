package webcrawler

import (
	"sync"
	"database"
	"models"
)

func Crawl(uri string, num int, indexer *database.Indexer) []*models.Document {
	pages := make([]*models.Document, 0)
	visited := make(map[string]bool)
	results := make(chan *models.Document)

	// the 'fetch' goroutine
	fetch := func(url string, wg *sync.WaitGroup) {
		page := Fetch(url)
		if page != nil {
			// send the page to results channel
			results <- page
			wg.Done()
		}
	}

	// the 'update' goroutine
	update := func(page *models.Document, wg *sync.WaitGroup) {
		indexer.UpdateOrAddPage(page)
		wg.Done()
	}

	var wg sync.WaitGroup
	wg.Add(num)
	go fetch(uri, &wg)
	visited[uri] = true

	for len(pages) < num {
		// append page from results channel
		page := <-results
		pages = append(pages, page)

		// update page to indexer
		wg.Add(1)
		go update(page, &wg)

		// fetch all unvisited links
		for _, link := range page.Links {
			if !visited[link] {
				go fetch(link, &wg)
				visited[link] = true
			}
		}
	}

	wg.Wait()
	return pages
}
