package controllers

import (
	"net/http"
	"html/template"
	"github.com/rsmohamad/comp4321/database"
	"log"
)

var homeTemplate = template.Must(template.ParseFiles("views/home.html"))

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr)
	userId := database.GetCookieInstance().GetCookieId(r)
	database.GetCookieInstance().SetCookieResponse(userId, w)
	homeTemplate.Execute(w, nil)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
}

func LoadHome(){
	homeTemplate = template.Must(template.ParseFiles("views/home.html"))
	staticServer := http.FileServer(http.Dir("static"))
	viewServer := http.FileServer(http.Dir("views"))

	http.Handle("/views/", http.StripPrefix("/views/", viewServer))
	http.Handle("/static/", http.StripPrefix("/static/", staticServer))
	http.HandleFunc("/", helloWorldHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)
}