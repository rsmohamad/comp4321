package stopword

import (
	"testing"
	"io/ioutil"
	"regexp"
	"strings"
)

func readStopwords() []string {
	data, err := ioutil.ReadFile("stopword/stopwords.txt")
	if err != nil {
		return nil
	}

	// Replace all newlines with '|'
	newlineRegex := regexp.MustCompile("\r?\n")
	stopWordsString := newlineRegex.ReplaceAllString(string(data), " ")
	stopWordsArr := strings.Split(stopWordsString, " ")

	return stopWordsArr
}

func TestIsStopWord(t *testing.T) {
	sw := readStopwords()

	for _, word := range sw {
		if !IsStopWord(word){
			t.Fail()
		}
	}

	if IsStopWord(" "){
		t.Fail()
	}

	if IsStopWord("adrian"){
		t.Fail()
	}
}