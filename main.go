package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

const URL_CHAR4 string = "http://char4.com/"

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

func getDomainsFromPage(content string, letterCount int) []string {
	regexStr := fmt.Sprintf("www\\.[a-z0-9]{%d}\\.com", letterCount)
	r, _ := regexp.Compile(regexStr)
	return r.FindAllString(content, -1)
}

const PATTERN_ALL_LETTERS string = "ALL LETTERS"
const PATTERN_ALL_NUMBERS string = "ALL NUMBERS"
const PATTERN_ABCD string = "ABCD"

func filter(domains []string, pattern string) (results []string) {
	var r *regexp.Regexp
	switch pattern {
	case PATTERN_ALL_LETTERS:
		r = regexp.MustCompile("www\\.[a-z]{4,}\\.com")
	case PATTERN_ALL_NUMBERS:
		r = regexp.MustCompile("www\\.\\d{4,}\\.com")
	}
	for i := 0; i < len(domains); i++ {
		if r.MatchString(domains[i]) {
			results = append(results, domains[i])
		}
	}
	return
}

func showResults(title string, domains []string) {
	fmt.Println(title + ":")
	for i := 0; i < len(domains); i++ {
		fmt.Println(domains[i])
	}
	if len(domains) == 0 {
		fmt.Println("none")
	}
	fmt.Println("")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":9094", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	pageContent, err := download(URL_CHAR4)
	if err != nil {
		fmt.Fprintf(w, "error:%s", err.Error())
		panic(err)
	}

	domains := getDomainsFromPage(pageContent, 4)

	fmt.Fprint(w, "All Letters:\n")
	letter_domains := filter(domains, PATTERN_ALL_LETTERS)
	for i := 0; i < len(letter_domains); i++ {
		fmt.Fprintf(w, "%s\n", letter_domains[i])
	}

	if len(letter_domains) == 0 {
		fmt.Fprint(w, "none")
	}

	fmt.Fprint(w, "------------------\n")

	fmt.Fprint(w, "All Numbers:\n")
	number_domains := filter(domains, PATTERN_ALL_NUMBERS)
	for i := 0; i < len(number_domains); i++ {
		fmt.Fprintf(w, "%s\n", letter_domains[i])
	}

	if len(number_domains) == 0 {
		fmt.Fprint(w, "none")
	}
}
