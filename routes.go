package main

import (
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

const CHAR4_DOMAIN_LENGTH int = len("www.abcd.com")
const CHAR5_DOMAIN_LENGTH int = len("www.abcde.com")

type KeywordDomains struct {
	Keyword string
	Domains []string
}

func handleError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// handle home page
func homeHanlder(w http.ResponseWriter, r *http.Request) {
	domains, err := getAllAvailableDomains()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	char4_domains := Filter(domains, func(domain string) bool {
		return len(domain) == CHAR4_DOMAIN_LENGTH
	})

	char5_domains := Filter(domains, func(domain string) bool {
		return len(domain) == CHAR5_DOMAIN_LENGTH
	})

	data := struct {
		Char4 []string
		Char5 []string
	}{
		char4_domains,
		char5_domains,
	}
	t, _ := template.ParseFiles("tmpls/index.tmpl")
	t.Execute(w, data)
}

func createWatcherHandler(w http.ResponseWriter, r *http.Request) {
	var errMessage string = ""
	if r.Method == "POST" {
		watcherName := strings.TrimSpace(r.FormValue("name"))
		patternStr := strings.TrimSpace(r.FormValue("patterns"))
		if patternStr != "" {
			patterns := strings.Split(patternStr, "\n")
			watcherID, err := addNewWatcher(watcherName)
			if err != nil {
				handleError(w, err)
				return
			}
			err = addOrUpdatePatterns(watcherID, patterns)
			if err != nil {
				handleError(w, err)
				return
			}
			http.Redirect(w, r, "/watch/"+watcherID, http.StatusTemporaryRedirect)
			return
		}
		errMessage = "Must have at least one pattern."
	}
	t, _ := template.ParseFiles("tmpls/new_watch.tmpl")
	t.Execute(w, errMessage)
}

func watcherHandler(w http.ResponseWriter, r *http.Request) {
	watchID := r.URL.Path[len("/watch/"):]
	if watchID == "new" {
		createWatcherHandler(w, r)
		return
	}

	name, err := getWatcherName(watchID)

	if err != nil {
		handleError(w, err)
		return
	}

	patterns, err := getPatterns(watchID)
	if err != nil {
		handleError(w, err)
		return
	}

	domains, err := getAllAvailableDomains()
	if err != nil {
		handleError(w, err)
		return
	}

	filterDomains := matchKeywords(domains, patterns)

	t, _ := template.ParseFiles("tmpls/watch.tmpl")
	t.Execute(w, struct {
		Name    string
		Domains []KeywordDomains
	}{
		name,
		filterDomains,
	})
}

// filter domains by pattern
func filterByPattern(domains []string, pattern string) (results []string) {
	r := regexp.MustCompile("www\\.[^.]*" + strings.TrimSpace(pattern) + "[^.]*\\.com")
	for _, domain := range domains {
		if r.MatchString(domain) {
			results = append(results, domain)
		}
	}
	return
}

// match domains to keywords
func matchKeywords(domains []string, patterns []string) (result []KeywordDomains) {
	for _, pattern := range patterns {
		result = append(result, KeywordDomains{pattern, filterByPattern(domains, pattern)})
	}
	return
}
