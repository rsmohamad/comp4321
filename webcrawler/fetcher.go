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
			newUrl := uri.Scheme + "://" + uri.Host + uri.Path
			rv = append(rv, newUrl)
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
		token = porter2.Stem(token)
		if len(token) > 1 && !stopword.IsStopWord(token) {
			rv = append(rv, token)
		}
	}
	return
}

var fetchClient = http.Client{Timeout: time.Second * 10}
func Fetch(uri string) (page *models.Document) {
	words := make([]string, 0)
	var lastElement string
	page = &models.Document{Uri: uri}
	inBody := false

	// Make HTTP GET request
	res, _ := fetchClient.Get(uri)

	// Return if HTTP request is not successful
	if res == nil || res.StatusCode != 200 {
		return nil
	}

	tm, _ := time.Parse(time.RFC1123, res.Header.Get("Last-Modified"))
	page.Modtime = tm.Unix()
	if page.Modtime < 0 {
		tm, _ = time.Parse(time.RFC1123, res.Header.Get("Date"))
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
				link, redirect := handleMetaRedirect(t)
				if redirect {
					return Fetch(link)
				}
				break
			}
			break
		case html.TextToken:
			// Skip if text is empty, not in between body tags or between script tags
			trimmed := strings.TrimSpace(t.Data)
			if trimmed != "" && inBody && lastElement != "script" && lastElement != "style" {
				words = append(words, trimmed)
			}
		}
	}

	// Clean data
	hue := tokenizeString(strings.Join(words, " "))
	// fmt.Println(hue)
	page.Titles = countTfandIdx(tokenizeString(page.Title))
	page.Words = countTfandIdx(hue)
	// fmt.Println(page.Words)
	page.MaxTf = countMaxTf(page.Words)
	page.Links = toAbsoluteUrl(page.Links, uri)
	return
}

func countTfandIdx(words []string) map[string]models.Word {
	m := make(map[string]models.Word)
	idx := 0
	for _, word := range words {
		count := m[word].Tf
		wordModel := m[word]
		wordModel.Tf = count + 1
		wordModel.Positions = append(wordModel.Positions, idx)
		m[word] = wordModel
		idx++
	}
	return m
}

func countMaxTf(words map[string]models.Word) int {
	max := 0
	for _, word := range words {
		if word.Tf > max {
			max = word.Tf
		}
	}
	return max
}
