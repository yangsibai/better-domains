package main

import (
	"fmt"
	"net/http"
)

func main() {
	fetchDomains()
	checkDomainRegister()

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", homeHanlder)
	http.HandleFunc("/watch/", watcherHandler)

	fmt.Println("server is listening")
	http.ListenAndServe(":9024", nil)
}
