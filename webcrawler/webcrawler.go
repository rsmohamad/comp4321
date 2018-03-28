package webcrawler

import (
	"comp4321/database"
	"comp4321/models"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
)

// HashMap for storing the robots data for each hostname
var robotMap sync.Map

func fetchRobotsTxt(urlObject *url.URL) (rv *robotstxt.RobotsData) {
	// HTTP client with 5 seconds timeout
	robotFetcher := http.Client{Timeout: time.Second * 5}

	// If there's no response or timeout exceeded, give the default robots checker
	// which will allow all paths within that website
	rv = &robotstxt.RobotsData{}
	robotUrl := fmt.Sprintf("%s://%s/robots.txt", urlObject.Scheme, urlObject.Host)
	res, _ := robotFetcher.Get(robotUrl)

	// Try parsing the robots.txt
	if res != nil && res.StatusCode == 200 {
		robots, err := robotstxt.FromResponse(res)
		if err == nil {
			rv = robots
			fmt.Println("Fetched " + robotUrl)
		}
		res.Body.Close()
	}
	return
}

// Checks if the url is crawlable.
// Will fetch robots.txt if not previously fetched.
// Thread safe.
func isAllowedToCrawl(link string) bool {
	// Use url object to process URLs
	urlObject, _ := url.Parse(link)
	var robots *robotstxt.RobotsData

	// Fetch a website's robots.txt if it's not already fetched
	res, found := robotMap.Load(urlObject.Host)
	if !found {
		// Ensure that the same robots.txt is not fetched twice
		robotMap.Store(urlObject.Host, nil)
		robots = fetchRobotsTxt(urlObject)
		robotMap.Store(urlObject.Host, robots)
	} else {
		// Wait for other thread to get the robots.txt
		for res == nil {
			res, found = robotMap.Load(urlObject.Host)
		}
		robots = res.(*robotstxt.RobotsData)
	}

	// Check with robots.txt
	if !robots.TestAgent(urlObject.Path, "Agent") {
		fmt.Println(urlObject.String() + " is not allowed; skipped!")
		return false
	}

	return true
}

// Concurrent routine for fetching a page.
// Feeds the page to results channel if fetch is successful.
// Feeds nil if fetch is unsuccessful.
var throttle = time.Tick(time.Millisecond * 10)
func concurrentFetch(url string, results *chan *models.Document, wg *sync.WaitGroup) {
	if !isAllowedToCrawl(url) {
		*results <- nil
		return
	}
	<-throttle
	page := Fetch(url)
	if page != nil && len(page.Words) > 0 && page.Title != "" {
		*results <- page
		wg.Done()
	} else {
		*results <- nil
	}
}

func Crawl(uri string, num int, index *database.Indexer) (pages []*models.Document) {
	var activeCounter int
	var fetchWg, updateWg sync.WaitGroup
	visited := make(map[string]bool)
	results := make(chan *models.Document)
	queue := make([]string, 0)

	// Sanitize url
	urlParse, _ := url.Parse(uri)
	queue = append(queue, urlParse.String())
	visited[urlParse.String()] = true

	fetchWg.Add(num)
	for len(pages) < num {
		// Create goroutines as needed
		needed := num - len(pages) - activeCounter
		for ; len(queue) > 0 && needed > 0; needed-- {
			activeCounter++
			go concurrentFetch(queue[0], &results, &fetchWg)
			queue = queue[1:]
		}

		// Retrieve one page from results channel
		page := <-results
		activeCounter--
		if page == nil {
			continue
		}

		pages = append(pages, page)
		fmt.Printf("Fetched page #%d out of %d : %s\n", len(pages), num, page.Uri)
		updateWg.Add(1)
		go func(i int, doc *models.Document) {
			index.UpdateOrAddPage(doc)
			//fmt.Printf("Indexed page #%d out of %d : %s\n", i, num, page.Uri)
			updateWg.Done()
		}(len(pages), page)

		// Put unvisited links into queue
		for _, link := range page.Links {
			// skip if the link is already visited or indexed
			if !visited[link] && !index.ContainsUrl(link) {
				queue = append(queue, link)
				visited[link] = true
			}
		}
	}

	fetchWg.Wait()
	updateWg.Wait()
	index.FlushInverted()
	return pages
}
