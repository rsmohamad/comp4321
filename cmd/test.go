package main

import (
	"bufio"
	"comp4321/database"
	"comp4321/models"
	"fmt"
	"os"
	"text/template"
)

func main() {
	viewer, _ := database.LoadViewer("index.db")
	file, _ := os.Create("spider_result.txt")
	fileStream := bufio.NewWriter(file)
	outTemplate, _ := template.ParseFiles("templates/testOutput.txt")

	defer file.Close()
	defer viewer.Close()

	viewer.ForEachDocument(func(p *models.Document, i int) {
		fmt.Println(i + 1)
		outTemplate.Execute(os.Stdout, p)
		outTemplate.Execute(fileStream, p)
		fileStream.Flush()
	})

}
