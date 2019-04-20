package main

import (
	"net/http"
	"net/url"
	"testing"
)

func TestGetRedirectGet(t *testing.T) {
	// Constructed URL to match a nginx redirect:
	// `return 302 https://login.luzifer.io/login?go=$scheme://$http_host$request_uri;`
	testURL := "https://example.com/login?go=https://example.com/inner?foo=bar&bar=foo"
	expectURL := "https://example.com/inner?bar=foo&foo=bar"

	req, _ := http.NewRequest(http.MethodGet, testURL, nil)

	rURL, err := getRedirectURL(req)
	if err != nil {
		t.Errorf("getRedirectURL caused an error in GET: %s", err)
	}

	if expectURL != rURL {
		t.Errorf("Result did not match expected URL: %q != %q", rURL, expectURL)
	}
}

func TestGetRedirectGetEmpty(t *testing.T) {
	testURL := "https://example.com/login"
	expectURL := ""

	req, _ := http.NewRequest(http.MethodGet, testURL, nil)

	rURL, err := getRedirectURL(req)
	if err != nil {
		t.Errorf("getRedirectURL caused an error in GET: %s", err)
	}

	if expectURL != rURL {
		t.Errorf("Result did not match expected URL: %q != %q", rURL, expectURL)
	}
}

func TestGetRedirectPost(t *testing.T) {
	testURL := "https://example.com/login"
	expectURL := "https://example.com/inner?foo=bar"

	body := url.Values{
		"go":    []string{expectURL},
		"other": []string{"param"},
	}
	req, _ := http.NewRequest(http.MethodPost, testURL, nil)
	req.Form = body // Force-set the form values to emulate parsed form

	rURL, err := getRedirectURL(req)
	if err != nil {
		t.Errorf("getRedirectURL caused an error in POST: %s", err)
	}

	if expectURL != rURL {
		t.Errorf("Result did not match expected URL: %q != %q", rURL, expectURL)
	}
}
