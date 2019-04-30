package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var ts = httptest.NewServer(http.HandlerFunc(dadataTestServerOk))
var secret = "12345"

type TestCase struct {
	url            string
	token          string
	xSecret        string
	expectedError  bool
	path           string
	testingReqOnly bool
}

var TestCases = []TestCase{
	{
		url:           ts.URL,
		token:         secret,
		xSecret:       secret,
		expectedError: false,
	},
	{
		url:            ts.URL,
		token:          secret,
		xSecret:        secret,
		expectedError:  false,
		path:           "profile/balance",
		testingReqOnly: true,
	},
	{
		url:            ts.URL,
		token:          secret,
		xSecret:        secret,
		path:           "unknown_path",
		testingReqOnly: true,
		expectedError:  true,
	},
	{
		url:            ts.URL,
		token:          secret,
		xSecret:        secret,
		path:           "bad_path\n",
		testingReqOnly: true,
		expectedError:  true,
	},
	{
		url:            "",
		token:          secret,
		xSecret:        secret,
		path:           "profile/balance",
		testingReqOnly: true,
		expectedError:  true,
	},
	{
		expectedError: true,
	},
	{
		url:           ts.URL,
		xSecret:       secret,
		expectedError: true,
	},
	{
		url:           ts.URL,
		token:         secret,
		expectedError: true,
	},
	{
		url:           ts.URL,
		token:         "bad_token",
		xSecret:       secret,
		expectedError: true,
	},
	{
		url: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})).URL,
		token:         secret,
		xSecret:       secret,
		expectedError: true,
	},
	{
		url: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{ "wrong": "json"`)
		})).URL,
		token:         secret,
		xSecret:       secret,
		expectedError: true,
	},
}

func dadataTestServerStats(w http.ResponseWriter, r *http.Request) {
	response := `
	{
    "date": "2018-09-12",
    "services": {
        "merging": 0,
        "suggestions": 11,
        "clean": 1004
    }
	}`
	fmt.Fprintf(w, response)
}

func dadataTestServerBalance(w http.ResponseWriter, r *http.Request) {
	response := `{ "balance": 9922.30 }`
	fmt.Fprintf(w, response)
}

func dadataTestServerOk(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != fmt.Sprintf("Token %s", secret) || r.Header.Get("X-Secret") != secret {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	switch r.URL.Path {
	case "/profile/balance":
		dadataTestServerBalance(w, r)
	case "/stat/daily":
		dadataTestServerStats(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func TestDadataRequestAdditional(t *testing.T) {
	c := Collector{}
	_, err := c.dadataRequest("bad_url")
	if err == nil {
		t.Errorf("Expected error")
	}
}

func TestDadata(t *testing.T) {
	for _, tc := range TestCases {
		c, err := NewCollector(tc.url, tc.token, tc.xSecret)
		if err != nil && !tc.expectedError {
			t.Errorf("Testcase: %v, error was not expected, got: %v", tc, err)
			continue
		} else if err != nil {
			continue
		}
		if tc.testingReqOnly {
			_, err := c.dadataRequest(tc.path)
			if err != nil && !tc.expectedError {
				t.Errorf("Error was not expected, got: %v", err)
			}
		}
		err = c.getDadataBalance()
		if err != nil && !tc.expectedError {
			t.Errorf("Error was not expected, got: %v", err)
		}
		err = c.getDadataStats()
		if err != nil && !tc.expectedError {
			t.Errorf("Error was not expected, got: %v", err)
		}
	}
}
