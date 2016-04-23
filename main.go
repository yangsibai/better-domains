package main

import (
	"net/http"
)

func main() {
	go fetchDomains()

	http.HandleFunc("/", homeHanlder)
	http.HandleFunc("/watch/", watcherHandler)
	http.ListenAndServe(":9024", nil)
}
