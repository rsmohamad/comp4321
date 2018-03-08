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

func searchHandler (w http.ResponseWriter, r *http.Request) {
	fmt.Println("Search " + r.URL.String())
}

func faviconHandler (w http.ResponseWriter, r *http.Request) {

}

func main(){
	// Handle requests for static files (CSS, JS)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static", http.StripPrefix("/static/", fs))

	// Handle requests for routes
	http.HandleFunc("/", helloWorldHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	fmt.Println(http.ListenAndServe(":80", nil))
}
