package main

import (
	"bytes"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

func index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func contains(vs []string, t string) bool {
	return index(vs, t) != -1
}

// sort and clean domains
func sortAndCleanDomains(domains []string) (results []string) {
	for _, domain := range domains {
		results = append(results, pureDomainName(domain))
	}
	return
}

// detect is domain registered
func isDomainRegistered(domain string) bool {
	return canDial(domain) || whoisQueryRegistered(domain)
}

func canDial(domain string) bool {
	if len(domain) <= 4 {
		log.Println("domain length is less than 4")
	}
	_, err := net.DialTimeout("tcp", trimHeadOfDomain(domain)+":80", time.Duration(config.DialTimeout)*time.Second)
	if err != nil {
		return false
	}
	return true
}

// use whois query to detect is domain registered
func whoisQueryRegistered(domain string) bool {
	cmd := exec.Command("sleep", "5")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Start()

	done := make(chan error, 1)

	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(time.Duration(config.WhoisQueryTimeout) * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			log.Printf("%s failed to kill process %v", domain, err)
			return false
		}
		log.Printf("%s process killed as timeout reached", domain)
		return false
	case err := <-done:
		if err != nil {
			return false
		} else {
			return out.String() != "" && strings.Index(out.String(), "No match for") == -1
		}
	}
}

// trim `www.` of a domain
func trimHeadOfDomain(domain string) string {
	if len(domain) > 4 {
		return domain[4:]
	}
	return domain
}

// get pure domain name, without `www.` and `.com`
func pureDomainName(domain string) string {
	if len(domain) > 4 && strings.Index(domain, ".com") > 4 {
		return domain[4:strings.Index(domain, ".com")]
	}
	return domain
}
