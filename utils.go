package main

import (
	"bytes"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

const dialTimeout time.Duration = 5 * time.Second
const whoisQueryTimeout time.Duration = 10 * time.Second

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
		if len(domain) > 4 && strings.Index(domain, ".com") > 4 {
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
	_, err := net.DialTimeout("tcp", domain[4:]+":80", dialTimeout)
	if err != nil {
		return false
	}
	return true
}

func whoisQueryRegistered(domain string) bool {
	cmd := exec.Command("whois", domain)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Start()

	done := make(chan error, 1)

	go func() {
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(whoisQueryTimeout):
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
			return strings.Index(out.String(), "No match for") == -1
		}
	}
}
