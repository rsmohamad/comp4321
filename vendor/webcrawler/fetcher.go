package webcrawler

import (
	"github.com/surgebase/porter2"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"models"
)

// Get url from html token
func getUrl(t html.Token) string {
	// Iterate over all attributes
	for _, a := range t.Attr {
		if a.Key == "href" {
			return a.Val
		}
	}
	return ""
}

// Fix relative URL to absolute URL
func toAbsoluteUrl(links []string, base string) (rv []string) {
	baseUrl, _ := url.Parse(base)
	for _, link := range links {
		uri, err := url.Parse(link)

		if err != nil {
			continue
		}

		if !uri.IsAbs() {
			uri = baseUrl.ResolveReference(uri)
		}

		if strings.HasPrefix(uri.String(), "http") {
			rv = append(rv, uri.String())
		}
	}
	return
}

// Clean and tokenize string
func tokenizeString(s string) (rv []string) {
	// Remove special chars
	regex := regexp.MustCompile("[^a-zA-Z0-9 ]")
	s = strings.ToLower(regex.ReplaceAllString(s, " "))

	// Split into tokens
	tokens := strings.Split(s, " ")

	for _, token := range tokens {

		// Exclude short words and stopwords
		token = strings.TrimSpace(token)
		if len(token) > 2 && !isStopWord(token) {
			rv = append(rv, porter2.Stem(token))
		}
	}
	return
}

func Fetch(uri string) (page *models.Document) {
	var builder strings.Builder
	var lastElement string
	page = &models.Document{Uri: uri}
	inBody := false

	// Make HTTP GET request
	res, _ := http.Get(uri)

	// Return if HTTP request is not successful
	if res == nil {
		return nil
	}
	if res.StatusCode != 200 {
		return nil
	}

	page.Len = res.ContentLength
	tm, _ := time.Parse(time.RFC1123, res.Header.Get("Last-Modified"))
	page.Modtime = tm.Unix()

	// Tokenize
	tokenizer := html.NewTokenizer(res.Body)
	defer res.Body.Close()

	// Loop through all html elements
	for {
		tokenType := tokenizer.Next()
		t := tokenizer.Token()
		if tokenType == html.ErrorToken {
			break
		}

		switch tokenType {
		case html.StartTagToken:
			// Indicate when inside body tags
			if t.Data == "body" {
				inBody = true
			}
			// Title
			if t.Data == "title" {
				tokenizer.Next()
				page.Title = strings.TrimSpace(tokenizer.Token().Data)
			}
			// Links
			if t.Data == "a" && inBody {
				page.Links = append(page.Links, getUrl(t))
			}
			lastElement = t.Data
			break

		case html.TextToken:
			// Skip if text is empty, not in between body tags or between script tags
			trimmed := strings.TrimSpace(t.Data)
			if trimmed != "" && inBody && lastElement != "script" {
				builder.WriteString(" " + trimmed)
			}
		}
	}

	// Clean data
	page.Words = countTf(tokenizeString(builder.String()))
	page.Links = toAbsoluteUrl(page.Links, uri)
	return
}

func countTf(words []string) map[string]int {
	m := make(map[string]int)

	for _, word := range words {
		count := m[word]
		m[word] = count + 1
	}

	return m
}
