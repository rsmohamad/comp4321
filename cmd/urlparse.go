package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"robotstxt"
	"sync"
	"time"
)

func main() {
	uri := "http://www.nytimes.com/"
	startRobot := time.Now()
	var res *http.Response
	var wg sync.WaitGroup
	var robots *robotstxt.RobotsData

	// Concurrently fetch robots.txt
	go func() {
		wg.Add(1)
		res, _ = http.Get(uri + "robots.txt")
	}()

	// Add 10 seconds timeout or till robots.txt is found
	for time.Since(startRobot).Nanoseconds() < 10000000000 && res == nil {
	}
	wg.Done()

	// If there's no response or timeout exceeded, break from program
	if res == nil || res.StatusCode != 200 {
		fmt.Println("Time since program ran: ", time.Since(startRobot))
		fmt.Println("Timeout exceeded or robots.txt not found")
		return
	} else {
		// Parse the robots.txt
		robots, _ = robotstxt.FromResponse(res)
		res.Body.Close()
	}

	// Check if url is allowed or not
	urlParse, _ := url.Parse(uri)

	if string(urlParse.Path[len(urlParse.Path)-1]) != "/" {
		urlParse.Path = urlParse.Path + "/" // Fix path
	}

	if !robots.TestAgent(urlParse.Path, "Agent") {
		log.Fatal("URI not allowed!")
		return
	}

	fmt.Println(urlParse.Path)
}
