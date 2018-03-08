package main

import (
	"fmt"
	"net/http"
	"html/template"
)

var t = template.Must(template.ParseFiles("views/home.html"))

func helloWorldHandler (w http.ResponseWriter, r *http.Request){
	t.Execute(w, nil)
	fmt.Println(r.URL.String())
}

func main(){
	// Handle requests for static files (CSS, JS)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Handle requests for routes
	http.HandleFunc("/", helloWorldHandler)
	fmt.Println(http.ListenAndServe(":80", nil))
}
