package main

import "testing"

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://github.com/foo/bar", true},
		{"http://github.com/foo/bar", true},
		{"https://example.com", true},
		{"http://example.com", true},
		{"github.com/foo/bar", false},
		{"ftp://example.com", false},
		{"", false},
		{"htt://example.com", false},
		{"httpx://example.com", false},
	}
	for _, tt := range tests {
		if got := isValidURL(tt.url); got != tt.want {
			t.Errorf("isValidURL(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}

func TestShouldShowHelp(t *testing.T) {
	tests := []struct {
		args []string
		want bool
	}{
		{[]string{"rinku"}, true},
		{[]string{"rinku", "--help"}, true},
		{[]string{"rinku", "-h"}, true},
		{[]string{"rinku", "scan"}, false},
		{[]string{"rinku", "convert"}, false},
		{[]string{"rinku", "migrate"}, false},
		{[]string{"rinku", "https://github.com/foo"}, false},
		{[]string{"rinku", "--help", "extra"}, false},
		{[]string{"rinku", "-h", "extra"}, false},
		{[]string{"rinku", "scan", "--help"}, false},
	}
	for _, tt := range tests {
		if got := shouldShowHelp(tt.args); got != tt.want {
			t.Errorf("shouldShowHelp(%v) = %v, want %v", tt.args, got, tt.want)
		}
	}
}
