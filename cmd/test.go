package main

import (
	"bufio"
	"comp4321/database"
	"comp4321/models"
	"fmt"
	"os"
	"html/template"
)

const text = "{{.Title}}\n{{.Uri}}\n{{.GetTimeStr}}, {{.GetSizeStr}}\n" +
	"{{range $key, $value := .Words}}{{$key}} {{$value}}; {{end}} {{range .Links}}\n" +
	"{{.}}{{end}}\n" +
	"-------------------------------------------------------------------------------------------\n"

func main() {
	viewer, _ := database.LoadViewer("index.db")
	file, _ := os.Create("spider_result.txt")
	fileStream := bufio.NewWriter(file)
	outTemplate := template.New("output_template")
	outTemplate.Parse(text)

	defer file.Close()
	defer viewer.Close()

	viewer.ForEachDocument(func(p *models.Document, i int) {
		fmt.Println(i + 1)
		outTemplate.Execute(os.Stdout, p)
		outTemplate.Execute(fileStream, p)
		fileStream.Flush()
	})
}
