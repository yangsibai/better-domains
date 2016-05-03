package main

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
	"strings"
	"time"
)

const driverName string = "mysql"
const dataSourceName string = "root:@/domains"

const domainStatusNew string = "NEW"
const domainStatusRegistered string = "REGISTERED"

const watcherStatusNew string = "NEW"
const watcherStatusDeleted string = "DELETED"

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
	db, err := sql.Open(driverName, dataSourceName)
	check(err)
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Domain (
			id int not null primary key auto_increment,
			domain varchar(13) not null,
			create_time datetime not null,
			update_time datetime,
			status text not null
		) engine = innodb default charset = utf8
		`)
	check(err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Watcher (
			id varchar(50) not null primary key,
			name text not null,
			create_time datetime not null,
			update_time datetime,
			status text not null
		) engine = innodb default charset = utf8
		`)
	check(err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Pattern (
			id int not null primary key auto_increment,
			watcherID text not null,
			name text not null,
			pattern text not null,
			create_time datetime not null,
			update_time datetime
		) engine = innodb default charset = utf8
		`)
	check(err)
}

// get all domains, including new, registered
func getAllDomains() (domains []string, err error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()
	rows, err := db.Query(`SELECT domain FROM Domain`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var domain string
		rows.Scan(&domain)
		domains = append(domains, domain)
	}
	return
}

// insert domains
func addDomains(domains []string) (err error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return
	}

	stmt, err := tx.Prepare(`
		INSERT INTO Domain(
			domain, create_time, status
		)
		VALUES (
			?, now(), ?
		)`)
	if err != nil {
		return
	}
	defer stmt.Close()

	existedDomains, err := getAllDomains()
	if err != nil {
		return
	}

	for _, domain := range domains {
		if !contains(existedDomains, domain) {
			_, err = stmt.Exec(domain, domainStatusNew)
			check(err)
		}
	}
	tx.Commit()
	return
}

func getADomainToCheck() (domain string, err error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()
	err = db.QueryRow(`
	SELECT domain
	FROM domain
	WHERE update_time IS NULL OR update_time < DATE_ADD(now(), INTERVAL -1 HOUR)
	ORDER BY update_time ASC
	LIMIT 1;
	`).Scan(&domain)
	return
}

// set domain has been registered
func updateDomainStatus(domain string, isRegistered bool) (err error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()
	var status string
	if isRegistered {
		status = domainStatusRegistered
	} else {
		status = domainStatusNew
	}

	_, err = db.Exec(`
		UPDATE Domain
		SET status = ?, update_time = now()
		WHERE domain = ?
		`, status, domain)
	return
}

// get all available domains
func getAllAvailableDomains() (domains []string, err error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT domain FROM Domain WHERE status = ?`, domainStatusNew)
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
func addNewWatcher(newID, name string) (watcherID string, err error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	if newID != "" {
		watcherID = newID
	} else {
		watcherID = makeID()
	}

	isExists, err := isWatcherIDExists(watcherID)
	if err != nil {
		return
	}
	if isExists {
		err = errors.New("address is exists")
		return
	}

	db.Exec(`
		INSERT INTO Watcher (
			id, name, create_time, status
		) VALUES (
			?, ?, now(), ?
		)`, watcherID, name, watcherStatusNew)
	return
}

func getWatcherName(watcherID string) (name string, err error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	err = db.QueryRow(`SELECT name FROM Watcher WHERE id = ?`, watcherID).Scan(&name)
	return
}

// detect is watcher ID exists
func isWatcherIDExists(watcherID string) (bool, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return false, err
	}
	defer db.Close()

	var exist int
	err = db.QueryRow(`SELECT EXISTS(SELECT ID FROM Watcher WHERE ID = ?) as exist`, watcherID).Scan(&exist)

	if err != nil {
		return false, err
	}
	return exist == 1, nil
}

// update watcher name
func updateWatcher(watcherID string, newWatcherID string, name string) error {
	if watcherID != newWatcherID {
		isExists, err := isWatcherIDExists(newWatcherID)
		if err != nil {
			return err
		}
		if isExists {
			return errors.New("address is exists")
		}
	}

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`UPDATE Watcher SET ID = ?, name = ? WHERE ID = ?`, newWatcherID, name, watcherID)
	return err
}

// add or update patterns
func addOrUpdatePatterns(watcherID string, patterns []string) (err error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	_, err = db.Exec(`DELETE FROM Pattern WHERE watcherID = ?`, watcherID)
	if err != nil {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		return
	}

	stmt, err := tx.Prepare(`
	INSERT INTO Pattern (
		watcherID, name, pattern, create_time
	)
	VALUES (
		?, ?, ?, now()
	)
	`)
	if err != nil {
		return
	}
	defer stmt.Close()
	for _, pattern := range patterns {
		_, err = stmt.Exec(watcherID, strings.TrimSpace(pattern), strings.TrimSpace(pattern))
		check(err)
	}
	tx.Commit()

	return
}

// get all patterns by watcherID
func getPatterns(watcherID string) (patterns []string, err error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT pattern FROM Pattern WHERE watcherID = ?`, watcherID)
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
