package webcrawler

import (
	"io/ioutil"
	"regexp"
)

var stopWordsRegex *regexp.Regexp

func loadStopWords() {
	data, err := ioutil.ReadFile("resources/stopwords.txt")
	if err != nil {
		return
	}

	// Replace all newlines with '|'
	newlineRegex := regexp.MustCompile("\r?\n")
	stopWords := newlineRegex.ReplaceAllString(string(data), "|")

	// Make regex
	stopWordsRegex = regexp.MustCompile("^(" + stopWords + ")$")
}

func isStopWord(s string) bool {
	// Check if string matches stopwords regex
	if stopWordsRegex == nil {
		loadStopWords()
	}
	return stopWordsRegex.MatchString(s)
}