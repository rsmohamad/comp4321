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

// HTTP client for fetching robots.txt, with 5 seconds timeout
var robotFetcher = http.Client{Timeout: time.Second * 5}

// HashMap for storing the robots data for each hostname
var robotMap sync.Map

func fetchRobotsTxt(urlObject *url.URL) (rv *robotstxt.RobotsData) {
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
		for res == nil{
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

func concurrentFetch(url string, results *chan *models.Document, wg *sync.WaitGroup) {
	if !isAllowedToCrawl(url) {
		return
	}
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
