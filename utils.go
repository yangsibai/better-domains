package main

import (
	"log"
	"net"
	"os/exec"
	"strings"
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

func sortAndCleanDomains(domains []string) (results []string) {
	for _, domain := range domains {
		if len(domain) > 4 {
			log.Println(domain)
			results = append(results, domain[4:strings.Index(domain, ".com")])
		}
	}
	return
}

func isDomainRegistered(domain string) bool {
	return canDial(domain) || whoisQueryRegistered(domain)
}

func canDial(domain string) bool {
	if len(domain) <= 4 {
		log.Println("domain length is less than 4")
	}
	_, err := net.Dial("tcp", domain[4:]+":80")
	if err != nil {
		return false
	}
	return true
}

func whoisQueryRegistered(domain string) bool {
	out, err := exec.Command("whois", domain).Output()
	if err != nil {
		log.Fatal(err)
		return false
	}
	return strings.Index(string(out), "No Match for") != -1
}
