package main

import (
	"net/http"
	"fmt"
)

func helloWorldHandler (w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "Hello World")
	fmt.Println(r.URL.String())
}

func main(){
	http.HandleFunc("/", helloWorldHandler)
	fmt.Println(http.ListenAndServe(":80", nil))
}
