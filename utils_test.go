package main

import "testing"

func TestCanDial(t *testing.T) {
	domain := "www.nobodyregisterthis.com"
	result := canDial(domain)
	if result {
		t.Error("Expected false, got", result)
	}
}

// this test case will failed in China because of the GFW
func TestCanDial2(t *testing.T) {
	domain := "www.google.com"
	result := canDial(domain)
	if !result {
		t.Error("Expected true, got", result)
	}
}

func TestWhoisQueryRegistered(t *testing.T) {
	domain := "www.nobodyregisterthis.com"
	result := whoisQueryRegistered(domain)
	if result {
		t.Error("Expected false, got", result)
	}
}

func TestWhoisQueryRegistered2(t *testing.T) {
	domain := "www.google.com"
	result := whoisQueryRegistered(domain)
	if !result {
		t.Error("Expected true, got", result)
	}
}
