package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const URL_CHAR4 string = "http://char4.com/"
const URL_CHAR5 string = "http://char5.com/"

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

func getDomainsFromPage(content string, charCount int) []string {
	regexStr := fmt.Sprintf("www\\.[a-z0-9]{%d}\\.com", charCount)
	r, _ := regexp.Compile(regexStr)
	return r.FindAllString(content, -1)
}

func fetchDomainsAndSave(url string, charCount int, saveTo string) {
	result, err := download(url)
	if err != nil {
		panic(err)
	}
	domains := getDomainsFromPage(result, charCount)
	err = ioutil.WriteFile(saveTo, []byte(strings.Join(domains, "\n")), 0644)
	if err != nil {
		panic(err)
	}
}

func fetchDomainsTick() {
	fetchDomainsAndSave(URL_CHAR4, 4, FILE_CHAR4)
	fetchDomainsAndSave(URL_CHAR5, 5, FILE_CHAR5)
}
