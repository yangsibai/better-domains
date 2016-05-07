package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type Configration struct {
	Port              string `json:"port"`
	WhoisQueryTimeout int    `json:"whoisQueryTimeout"`
	DialTimeout       int    `json:"dialTimeout"`
	Redis             struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		DB       int64  `json:"db"`
	}
}

var config Configration

func main() {
	go fetchDomains()
	go checkDomainRegister()

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	http.HandleFunc("/", homeHanlder)
	http.HandleFunc("/watch/", watcherHandler)

	log.Println("server is listening at", config.Port)
	http.ListenAndServe(":"+config.Port, nil)
}

func init() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("read config.json fail", err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("decode config.json fail", err)
	}
}
