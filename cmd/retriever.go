package main

import (
	"comp4321/database"
	"comp4321/webcrawler"
	"strings"

	"github.com/boltdb/bolt"
)

func main() {
	query := "computer business"
	querySplit := strings.Split(query, " ")

	preprocessText(&querySplit)

	queryID := convertWordsToID(&querySplit)

	db, err := bolt.Open("index.db", 0600, nil)
	if err != nil {
		fatal(err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) {
		invertedTable, _ := tx.Bucket([]byte(database.InvertedTable))
		return nil
	})

}

func preprocessText(words *[]string) {
	for i, word := range *words {
		if webcrawler.isStopWord(word) {
			*words = append(*words[:i], *words[i+1:]...)
		}
	}
}

func convertWordsToID(words *[]string) (idArr [][]byte) {
	for _, word := range *words {
		idArr = append(idArr, database.getWordId(word))
	}
}
