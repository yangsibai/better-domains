package main

import "testing"

func TestCanDial(t *testing.T) {
	domain := "www.nobodyregisterthis.com"
	result := canDial(domain)
	if result {
		t.Error("Expected false, got", result)
	}
}

func TestCanDial2(t *testing.T) {
	domain := "www.example.com"
	result := canDial(domain)
	if !result {
		t.Error("Expected true, got", result)
	}
}
