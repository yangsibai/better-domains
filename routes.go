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

type WatcherFormValue struct {
	Name     string
	Patterns []string
}

// read form value
func readForm(r *http.Request) WatcherFormValue {
	patternStr := strings.TrimSpace(r.FormValue("patterns"))
	if patternStr != "" {
		return WatcherFormValue{
			r.FormValue("name"),
			strings.Split(patternStr, "\n"),
		}
	}
	return WatcherFormValue{
		r.FormValue("name"),
		nil,
	}
}

// handle server error
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

// create watcher handler
func createWatcherHandler(w http.ResponseWriter, r *http.Request) {
	var errMessage string = ""
	if r.Method == "POST" {
		form := readForm(r)
		if len(form.Patterns) > 0 {
			watcherID, err := addNewWatcher(form.Name)
			if err != nil {
				handleError(w, err)
				return
			}
			err = addOrUpdatePatterns(watcherID, form.Patterns)
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

// update watcher
func editWatcherHandler(w http.ResponseWriter, r *http.Request) {
	watcherID := r.URL.Path[len("/watch/edit/"):]
	var errMessage string = ""
	var name string

	t, _ := template.ParseFiles("tmpls/edit_watch.tmpl")

	if r.Method == "POST" {
		form := readForm(r)
		err := updateWatcherName(watcherID, form.Name)
		if err != nil {
			handleError(w, err)
			return
		}

		err = addOrUpdatePatterns(watcherID, form.Patterns)
		if err != nil {
			handleError(w, err)
			return
		}

		http.Redirect(w, r, "/watch/"+watcherID, http.StatusTemporaryRedirect)

		return
	}
	name, err := getWatcherName(watcherID)
	if err != nil {
		handleError(w, err)
		return
	}

	patterns, err := getPatterns(watcherID)
	if err != nil {
		handleError(w, err)
		return
	}
	t.Execute(w, struct {
		WatcherID string
		Name      string
		Error     string
		Pattern   string
	}{
		watcherID,
		name,
		errMessage,
		strings.Join(patterns, "\n"),
	})
}

func watcherHandler(w http.ResponseWriter, r *http.Request) {
	watchID := r.URL.Path[len("/watch/"):]
	if watchID == "new" {
		createWatcherHandler(w, r)
		return
	} else if strings.Index(watchID, "edit") != -1 {
		editWatcherHandler(w, r)
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
		WatcherID string
		Name      string
		Domains   []KeywordDomains
	}{
		watchID,
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
