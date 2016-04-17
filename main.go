package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

const URL_CHAR4 string = "http://char4.com/"
const URL_CHAR5 string = "http://char5.com/"
const PATTERN_ALL_LETTERS string = "ALL LETTERS"
const PATTERN_ALL_NUMBERS string = "ALL NUMBERS"
const PATTERN_ABCD string = "ABCD"

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

func filter(domains []string, pattern string) (results []string) {
	var r *regexp.Regexp
	switch pattern {
	case PATTERN_ALL_LETTERS:
		r = regexp.MustCompile("www\\.[a-z]{4,}\\.com")
	case PATTERN_ALL_NUMBERS:
		r = regexp.MustCompile("www\\.\\d{4,}\\.com")
	default:
		r = regexp.MustCompile(pattern)
	}
	for i := 0; i < len(domains); i++ {
		if r.MatchString(domains[i]) {
			results = append(results, domains[i])
		}
	}
	return
}

func findAndListDomains(w http.ResponseWriter, domains []string, title string, pattern string) {
	fmt.Fprintf(w, "%s:\n", title)
	sub_domains := filter(domains, pattern)
	for i := 0; i < len(sub_domains); i++ {
		fmt.Fprintf(w, "%s\n", sub_domains[i])
	}
	if len(sub_domains) == 0 {
		fmt.Fprint(w, "none\n")
	}
	fmt.Fprint(w, "\n")
}

func listFromURL(w http.ResponseWriter, domains []string, title string) {
	fmt.Fprintf(w, "%s\n\n", title)

	findAndListDomains(w, domains, "All Letters", PATTERN_ALL_LETTERS)

	findAndListDomains(w, domains, "All Numbers", PATTERN_ALL_NUMBERS)

	findAndListDomains(w, domains, "Contain si", "si")

	findAndListDomains(w, domains, "Contain bai", "bai")

	findAndListDomains(w, domains, "Contain bo", "bo")

	findAndListDomains(w, domains, "Contain qq", "qq")

	findAndListDomains(w, domains, "Contain go", "go")
}

func getDomainsFromURL(url string, count int) []string {
	pageContent, err := download(url)
	if err != nil {
		panic(err)
	}

	domains := getDomainsFromPage(pageContent, count)
	return domains
}

func handler(w http.ResponseWriter, r *http.Request) {
	char4_domains := getDomainsFromURL(URL_CHAR4, 4)
	listFromURL(w, char4_domains, "4 Characters")

	char5_domains := getDomainsFromURL(URL_CHAR5, 5)
	listFromURL(w, char5_domains, "5 Characters")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":9024", nil)
}
