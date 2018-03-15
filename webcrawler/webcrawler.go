package webcrawler

import (
	"comp4321/database"
	"comp4321/models"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"robotstxt"
	"sync"
	"time"
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
	startRobot := time.Now()
	var res *http.Response
	var wg sync.WaitGroup
	var robots *robotstxt.RobotsData

	// Check if url is allowed or not
	urlParse, _ := url.Parse(uri)

	if len(urlParse.Path) == 0 || string(urlParse.Path[len(urlParse.Path)-1]) != "/" {
		urlParse.Path = urlParse.Path + "/" // Fix path
	}

	fmt.Println(urlParse.Path)

	// Concurrently fetch robots.txt
	go func() {
		wg.Add(1)
		res, _ = http.Get(urlParse.String() + "robots.txt")
	}()

	// Add 10 seconds timeout or till robots.txt is found
	for time.Since(startRobot).Nanoseconds() < 10000000000 && res == nil {
	}
	wg.Done()

	// If there's no response or timeout exceeded, break from program
	if res == nil || res.StatusCode != 200 {
		fmt.Println("Time since program ran: ", time.Since(startRobot))
		fmt.Println("Timeout exceeded or robots.txt not found")
		robots = nil
	} else {
		// Parse the robots.txt
		robots, _ = robotstxt.FromResponse(res)
		res.Body.Close()
	}

	pages := make([]*models.Document, 0)
	visited := make(map[string]bool)
	results := make(chan *models.Document)
	fetchWg := sync.WaitGroup{}
	updateWg := sync.WaitGroup{}
	fetchWg.Add(num)

	if robots != nil && !robots.TestAgent(urlParse.Path, "Agent") {
		log.Fatal("URI not allowed!")
		return nil
	}

	// Visit first page
	go concurrentFetch(uri, &results, &fetchWg)
	visited[uri] = true

	for len(pages) < num {
		// append page from results channel
		page := <-results
		pages = append(pages, page)

		fmt.Printf("Fetched page #%d out of %d : %s\n", len(pages), num, page.Uri)

		updateWg.Add(1)
		go func(i int) {
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
				// Check if url is allowed or not
				urlParse, _ := url.Parse(link)

				if string(urlParse.Path[len(urlParse.Path)-1]) != "/" {
					urlParse.Path = urlParse.Path + "/" // Fix path
				}

				if string(urlParse.Path[0]) != "/" {
					urlParse.Path = "/" + urlParse.Path // Fix path
				}

				if robots != nil && !robots.TestAgent(urlParse.Path, "Agent") {
					fmt.Println(link + "not allowed; skipped!")
					continue
				}

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
