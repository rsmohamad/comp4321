package main

import (
	"fmt"
	"log"
	"net/http"
	"robotstxt"
	"sync"
	"time"
)

func main() {
	start := time.Now()
	var res *http.Response
	var wg sync.WaitGroup

	// Concurrently fetch robots.txt
	go func() {
		wg.Add(1)
		res, _ = http.Get("http://www.nytimes.com/robots.txt")
	}()

	// Add 10 seconds timeout or till robots.txt is found
	for time.Since(start).Nanoseconds() < 10000000000 && res == nil {
	}
	wg.Done()

	// If there's no response or timeout exceeded, break from program
	if res == nil || res.StatusCode != 200 {
		fmt.Println("Time since program ran: ", time.Since(start))
		fmt.Println("Timeout exceeded or robots.txt not found")
		return
	}

	// Parse the robots.txt
	robots, err := robotstxt.FromResponse(res)
	res.Body.Close()
	if err != nil {
		log.Println("Error parsing robots.txt:", err.Error())
	}

	// Test if url is allowed or not
	allow := robots.TestAgent("/paidcontent/", "Agent")
	fmt.Println(allow)
}
