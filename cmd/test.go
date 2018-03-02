package main

import (
	"models"
	"database"
	"text/template"
	"os"
	"bufio"
	"fmt"
)

func main() {
	index, _ := database.LoadIndexer("index.db")
	file, _ := os.Create("spider_test.txt")
	fileStream := bufio.NewWriter(file)
	outTemplate, _ := template.ParseFiles("templates/testOutput.txt")

	defer file.Close()
	defer index.Close()

	index.ForEachDocument(func(p *models.Document, i int) {
		fmt.Println(i + 1)
		outTemplate.Execute(os.Stdout, p)
		outTemplate.Execute(fileStream, p)
		fileStream.Flush()
	})

}
