package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const PATTERN_ALL_LETTERS string = "ALL LETTERS"
const PATTERN_ALL_NUMBERS string = "ALL NUMBERS"
const PATTERN_ABCD string = "ABCD"

const FILE_CHAR4 string = "char4.txt"
const FILE_CHAR5 string = "char5.txt"

var KEYWORDS []string

type Domains struct {
	Domains []string
	Keyword []KeywordDomains
}

type KeywordDomains struct {
	Keyword string
	Domains []string
}

// filter domains by pattern
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

// read domains from file
func getDomainsFromFile(filename string) (domains []string, err error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	content := strings.Trim(string(bs), "\n")
	domains = strings.Split(content, "\n")
	return
}

// match domains to keywords
func matchKeywords(domains []string) (result []KeywordDomains) {
	result = append(result, KeywordDomains{"All Letters", filter(domains, PATTERN_ALL_LETTERS)})
	result = append(result, KeywordDomains{"All Numbers", filter(domains, PATTERN_ALL_NUMBERS)})
	for i := 0; i < len(KEYWORDS); i++ {
		result = append(result, KeywordDomains{KEYWORDS[i], filter(domains, KEYWORDS[i])})
	}
	return
}

// handle home page
func homeHanlder(w http.ResponseWriter, r *http.Request) {
	char4_domains, err := getDomainsFromFile(FILE_CHAR4)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	char5_domains, err := getDomainsFromFile(FILE_CHAR5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Char4 Domains
		Char5 Domains
	}{
		Domains{char4_domains, matchKeywords(char4_domains)},
		Domains{char5_domains, matchKeywords(char5_domains)},
	}
	t, _ := template.ParseFiles("tmpls/index.tmpl")
	t.Execute(w, data)
}

func main() {
	http.HandleFunc("/", homeHanlder)
	http.ListenAndServe(":9024", nil)
	fmt.Println("test")
}

func init() {
	bs, err := ioutil.ReadFile("keywords.txt")
	if err != nil {
		panic(err)
	}
	content := strings.Trim(string(bs), "\n")
	KEYWORDS = strings.Split(content, "\n")
}
