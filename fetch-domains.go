package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

const URL_CHAR4 string = "http://char4.com/"
const URL_CHAR5 string = "http://char5.com/"

var IGNORE_DOMAINS = []string{"www.char3.com", "www.char4.com", "www.char5.com"}

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
	return Index(IGNORE_DOMAINS, domain) == -1
}

func getDomainsFromPage(content string, charCount int) []string {
	regexStr := fmt.Sprintf("www\\.[a-z0-9]{%d}\\.com", charCount)
	r, _ := regexp.Compile(regexStr)
	domains := r.FindAllString(content, -1)
	if charCount == 5 {
		return Filter(domains, validDomain)
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
	fetchDomainsAndSave(URL_CHAR4, 4)
	fetchDomainsAndSave(URL_CHAR5, 5)
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
