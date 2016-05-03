package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
)

const char4URL string = "http://char4.com/"
const char5URL string = "http://char5.com/"

var domainsIgnore = []string{"www.char3.com", "www.char4.com", "www.char5.com"}

// download content from URL
func download(url string) (result string, err error) {
	response, err := http.Get(url)
	if err != nil {
		return
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	result = string(contents)
	return
}

func validDomain(domain string) bool {
	return index(domainsIgnore, domain) == -1
}

func getDomainsFromPage(content string, charCount int) []string {
	regexStr := fmt.Sprintf("www\\.[a-z0-9]{%d}\\.com", charCount)
	r, _ := regexp.Compile(regexStr)
	domains := r.FindAllString(content, -1)
	if charCount == 5 {
		return filter(domains, validDomain)
	}
	return domains
}

func fetchDomainsAndSave(url string, charCount int) {
	result, err := download(url)
	if err != nil {
		panic(err)
	}
	domains := getDomainsFromPage(result, charCount)
	err = addDomains(domains)
	if err != nil {
		panic(err)
	}
}

func fetchDomainsTick() {
	fetchDomainsAndSave(char4URL, 4)
	fetchDomainsAndSave(char5URL, 5)
}

func fetchDomains() {
	ticker := time.NewTicker(time.Minute * 10)
	fetchDomainsTick() // fetch once
	go func() {
		for range ticker.C {
			fetchDomainsTick()
		}
	}()
}

func getDomain(dc chan string) {
	for {
		domain, err := getADomainToCheck()
		if err != nil {
			dc <- ""
		}
		dc <- domain
	}
}

func checkDomain(dc chan string) {
	for {
		select {
		case domain := <-dc:
			registered := isDomainRegistered(domain)
			updateDomainStatus(domain, registered)
			log.Printf("%s registered: %v", domain, registered)
			time.Sleep(3 * 1e9)
		}
	}
}

func checkDomainRegister() {
	dc := make(chan string)
	go getDomain(dc)
	go checkDomain(dc)
}
