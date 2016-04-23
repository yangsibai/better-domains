package main

import (
	"net/http"
)

func main() {
	go fetchDomains()

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", homeHanlder)
	http.HandleFunc("/watch/", watcherHandler)

	http.ListenAndServe(":9024", nil)
}
