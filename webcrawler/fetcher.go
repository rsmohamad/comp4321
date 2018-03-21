package webcrawler

import (
	"comp4321/models"
	"comp4321/stopword"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/surgebase/porter2"
	"golang.org/x/net/html"
	"mvdan.cc/xurls"
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

func getMetaDesc(t html.Token) (desc string, found bool) {
	found = false
	for _, a := range t.Attr {
		if a.Key == "name" && a.Val == "description" {
			found = true
		}
		if a.Key == "content" && found {
			desc = a.Val
		}
	}
	return
}

func handleMetaRedirect(t html.Token) (url string, redirect bool) {
	redirect = false
	for _, a := range t.Attr {
		if a.Key == "http-equiv" && a.Val == "refresh" {
			redirect = true
		}
		if a.Key == "content" && redirect {
			url = xurls.Strict().FindString(a.Val)
		}
	}
	return
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
		if len(token) > 2 && !stopword.IsStopWord(token) {
			var t = ""
			if token == "hong" || token == "kong" {
				t = "hong kong"
			} else {
				t = porter2.Stem(token)
			}
			rv = append(rv, t)
		}
	}
	return
}

func Fetch(uri string) (page *models.Document) {
	words := make([]string, 0)
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

	tm, _ := time.Parse(time.RFC1123, res.Header.Get("Last-Modified"))
	page.Modtime = tm.Unix()
	if page.Modtime < 0 {
		tm, _ := time.Parse(time.RFC1123, res.Header.Get("Date"))
		page.Modtime = tm.Unix()
	}
	page.Len = 0

	// Tokenize
	tokenizer := html.NewTokenizer(res.Body)
	defer res.Body.Close()

	// Loop through all html elements
	for {
		tokenType := tokenizer.Next()
		t := tokenizer.Token()

		page.Len += len(tokenizer.Raw())

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
		case html.SelfClosingTagToken:
			if t.Data == "meta" {
				desc, found := getMetaDesc(t)
				if found {
					page.Desc = desc
				}
				link, redirect := handleMetaRedirect(t)
				if redirect {
					return Fetch(link)
				}
				break
			}

		case html.TextToken:
			// Skip if text is empty, not in between body tags or between script tags
			trimmed := strings.TrimSpace(t.Data)
			if trimmed != "" && inBody && lastElement != "script" {
				words = append(words, trimmed)
			}
		}
	}

	// Clean data
	page.Titles = countTf(tokenizeString(page.Title))
	page.Words = countTf(tokenizeString(strings.Join(words, " ")))
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
