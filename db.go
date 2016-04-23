package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strconv"
	"time"
)

const ENGINE string = "sqlite3"
const DATABASE string = "domain.db"

const DOMAIN_STATUS_NEW string = "NEW"
const DOMAIN_STATUS_REGISTERED string = "REGISTERED"

const WATCHER_STATUS_NEW string = "NEW"
const WATCHER_STATUS_DELETED string = "DELETED"

type Pattern struct {
	name    string
	pattern string
}

// sqls
// sql for domains
const SQL_CREATE_TABLE_DOMAIN = `
	CREATE TABLE IF NOT EXISTS Domain (
		id text not null primary key,
		domain text not null,
		create_time timestamp not null,
		update_time timestamp,
		status text not null
	)`
const SQL_QUERY_DOMAINS = "SELECT domain FROM Domain WHERE status = ?"
const SQL_INSERT_NEW_DOMAIN string = `
	INSERT INTO Domain(id, domain, create_time, status)
	VALUES (?, ?, ?, ?)`
const SQL_UPDATE_DOMAIN string = `
	UPDATE Domain
	SET status = ?
	WHERE id = ?`
const SQL_QUERY_BY_DOMAIN = "SELECT * FROM Domain WHERE domain = ?"

// sql for watchers
const SQL_CREATE_TABLE_WATCHER string = `
	CREATE TABLE IF NOT EXISTS Watcher (
		id text not null primary key,
		create_time timestamp not null,
		update_time timestamp,
		status text not null
	)`
const SQL_INSERT_NEW_WATCHER string = "INSERT INTO Watcher (id, create_time, status) VALUES (?, ?, ?)"

// sql for patterns
const SQL_CREATE_TABLE_PATTERN = `
	CREATE TABLE IF NOT EXISTS Pattern (
		id text not null primary key,
		watcherID text not null,
		name text not null,
		pattern text not null,
		create_time timestamp not null,
		update_time timestamp
	)`
const SQL_DELETE_PATTERS string = "DELETE FROM Pattern WHERE watcherID = ?"
const SQL_INSERT_NEW_PATTERN string = `
	INSERT INTO Pattern (
		id, watcherID, name, pattern, create_time
	)
	VALUES (
		?, ?, ?, ?, ?
	)`
const SQL_QUERY_PATTERNS string = "SELECT pattern FROM Pattern WHERE watcherID = ?"

type Domain struct {
	id          string
	domain      string
	create_time time.Time
	update_time time.Time
}

// check error, if error is not nil, panic
func check(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

// make an ID
func makeID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// initial database if need
func initDatabaseIfNeed() {
	db, err := sql.Open(ENGINE, DATABASE)
	check(err)
	defer db.Close()

	_, err = db.Exec(SQL_CREATE_TABLE_DOMAIN)
	check(err)

	_, err = db.Exec(SQL_CREATE_TABLE_WATCHER)
	check(err)

	sqlCreateTablePattern := SQL_CREATE_TABLE_PATTERN
	_, err = db.Exec(sqlCreateTablePattern)
	check(err)
}

// detect domain has exist
func exist(domain string) bool {
	db, err := sql.Open(ENGINE, DATABASE)
	check(err)
	defer db.Close()

	rows, err := db.Query(SQL_QUERY_BY_DOMAIN, domain)
	check(err)

	return rows.Next()
}

// insert domains
func addDomains(domains []string) (err error) {
	db, err := sql.Open(ENGINE, DATABASE)
	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return
	}

	stmt, err := tx.Prepare(SQL_INSERT_NEW_DOMAIN)
	if err != nil {
		return
	}
	defer stmt.Close()
	for _, domain := range domains {
		if !exist(domain) {
			_, err = stmt.Exec(makeID(), domain, time.Now(), DOMAIN_STATUS_NEW)
			check(err)
		}
	}
	tx.Commit()
	return
}

// set domain has been registered
func updateDomainAsRegistered(domain string) (err error) {
	db, err := sql.Open(ENGINE, DATABASE)
	if err != nil {
		return
	}
	defer db.Close()
	_, err = db.Exec(SQL_UPDATE_DOMAIN, DOMAIN_STATUS_REGISTERED, domain)
	return
}

// get all available domains
func getAllAvailableDomains() (domains []string, err error) {
	db, err := sql.Open(ENGINE, DATABASE)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query(SQL_QUERY_DOMAINS, DOMAIN_STATUS_NEW)
	if err != nil {
		return
	}

	for rows.Next() {
		var domain string
		err = rows.Scan(&domain)
		domains = append(domains, domain)
		check(err)
	}
	return
}

// create new watcher
func addNewWatcher() (watcherID string, err error) {
	db, err := sql.Open(ENGINE, DATABASE)
	if err != nil {
		return
	}
	defer db.Close()
	watcherID = makeID()
	db.Exec(SQL_INSERT_NEW_WATCHER, watcherID, time.Now(), WATCHER_STATUS_NEW)
	return
}

// add or update patterns
func addOrUpdatePatterns(watcherID string, patterns []string) (err error) {
	db, err := sql.Open(ENGINE, DATABASE)
	if err != nil {
		return
	}
	defer db.Close()
	_, err = db.Exec(SQL_DELETE_PATTERS, watcherID)
	if err != nil {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		return
	}

	stmt, err := tx.Prepare(SQL_INSERT_NEW_PATTERN)
	if err != nil {
		return
	}
	defer stmt.Close()
	for _, pattern := range patterns {
		_, err = stmt.Exec(makeID(), watcherID, pattern, pattern, time.Now())
		check(err)
	}
	tx.Commit()

	return
}

// get all patterns by watcherID
func getPatterns(watcherID string) (patterns []string, err error) {
	db, err := sql.Open(ENGINE, DATABASE)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query(SQL_QUERY_PATTERNS, watcherID)
	if err != nil {
		return
	}

	for rows.Next() {
		var pattern string
		rows.Scan(&pattern)
		patterns = append(patterns, pattern)
	}
	return
}

func init() {
	initDatabaseIfNeed()
}
