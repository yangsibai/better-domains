package main

import (
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

const char4DomainLength int = len("www.abcd.com")
const char5DomainLength int = len("www.abcde.com")

type keywordDomains struct {
	Keyword string
	Domains []string
}

type watcherFormValue struct {
	ID       string
	Name     string
	Patterns []string
}

type watcherPage struct {
	Button    string
	PageTitle string
	SubmitURL string
	WatcherID string
	Name      string
	Error     string
	Pattern   string
}

// read form value
func readForm(r *http.Request) watcherFormValue {
	patternStr := strings.TrimSpace(r.FormValue("patterns"))
	if patternStr != "" {
		return watcherFormValue{
			r.FormValue("ID"),
			r.FormValue("name"),
			strings.Split(patternStr, "\n"),
		}
	}
	return watcherFormValue{
		r.FormValue("ID"),
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
	char4Domains := filter(domains, func(domain string) bool {
		return len(domain) == char4DomainLength
	})

	char5Domains := filter(domains, func(domain string) bool {
		return len(domain) == char5DomainLength
	})

	data := struct {
		Char4 []string
		Char5 []string
	}{
		char4Domains,
		char5Domains,
	}
	t, _ := template.ParseFiles("tmpls/index.tmpl")
	t.Execute(w, data)
}

func renderWatcherEditOrAdd(w http.ResponseWriter, isAdd bool, page watcherPage) {
	if isAdd {
		page.SubmitURL = "/watch/new"
		page.PageTitle = "Add new watch list"
		page.Button = "Create"
	} else {
		page.SubmitURL = "/watch/edit/" + page.WatcherID
		page.PageTitle = "Edit watch list"
		page.Button = "Update"
	}
	t, _ := template.ParseFiles("tmpls/edit_watch.tmpl")
	t.Execute(w, page)
}

// create watcher handler
func createWatcherHandler(w http.ResponseWriter, r *http.Request) {
	var pageValue watcherPage
	if r.Method == "POST" {
		form := readForm(r)
		pageValue.Name = form.Name
		pageValue.Pattern = strings.Join(form.Patterns, "\n")
		pageValue.WatcherID = form.ID

		if len(form.Patterns) == 0 {
			pageValue.Error = "Must have at least one pattern."
			renderWatcherEditOrAdd(w, true, pageValue)
			return
		}

		watcherID, err := addNewWatcher(form.ID, form.Name)
		if err != nil {
			pageValue.Error = err.Error()
			renderWatcherEditOrAdd(w, true, pageValue)
			return
		}
		err = addOrUpdatePatterns(watcherID, form.Patterns)
		if err != nil {
			pageValue.Error = err.Error()
			renderWatcherEditOrAdd(w, true, pageValue)
			return
		}
		http.Redirect(w, r, "/watch/"+watcherID, http.StatusTemporaryRedirect)
		return
	}
	renderWatcherEditOrAdd(w, true, pageValue)
}

// update watcher
func editWatcherHandler(w http.ResponseWriter, r *http.Request) {
	watcherID := r.URL.Path[len("/watch/edit/"):]
	var page watcherPage
	page.WatcherID = watcherID

	if r.Method == "POST" {
		form := readForm(r)
		page.Name = form.Name
		page.Pattern = strings.Join(form.Patterns, "\n")
		page.WatcherID = form.ID

		err := updateWatcher(watcherID, form.ID, form.Name)
		if err != nil {
			page.Error = err.Error()
			renderWatcherEditOrAdd(w, false, page)
			return
		}

		err = addOrUpdatePatterns(form.ID, form.Patterns)
		if err != nil {
			page.Error = err.Error()
			renderWatcherEditOrAdd(w, false, page)
			return
		}

		http.Redirect(w, r, "/watch/"+form.ID, http.StatusTemporaryRedirect)
		return
	}

	name, err := getWatcherName(watcherID)
	if err != nil {
		page.Error = err.Error()
		renderWatcherEditOrAdd(w, false, page)
		return
	}
	page.Name = name

	patterns, err := getPatterns(watcherID)
	if err != nil {
		page.Error = err.Error()
		renderWatcherEditOrAdd(w, false, page)
		return
	}
	page.Pattern = strings.Join(patterns, "\n")

	renderWatcherEditOrAdd(w, false, page)
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
		Domains   []keywordDomains
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
func matchKeywords(domains []string, patterns []string) (result []keywordDomains) {
	for _, pattern := range patterns {
		result = append(result, keywordDomains{pattern, filterByPattern(domains, pattern)})
	}
	return
}
