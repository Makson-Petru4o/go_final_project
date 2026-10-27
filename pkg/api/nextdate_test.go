package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNextDateRules(t *testing.T) {
	cases := []struct {
		name, now, start, rule, want string
	}{
		{"daily sequence", "20240126", "20240113", "d 7", "20240127"},
		{"leap day", "20240228", "20240228", "d 1", "20240229"},
		{"leap year normalization", "20240126", "20240229", "y", "20250301"},
		{"yearly sequence", "20280228", "20240229", "y", "20280301"},
		{"month boundary", "20240430", "20240430", "d 1", "20240501"},
		{"year boundary", "20241231", "20241231", "d 1", "20250101"},
		{"sunday", "20240126", "20240126", "w 7", "20240128"},
		{"last leap day", "20240201", "20240101", "m -1", "20240229"},
		{"penultimate leap day", "20240201", "20240101", "m -2", "20240228"},
		{"last ordinary day", "20230201", "20230101", "m -1", "20230228"},
		{"month filter", "20240201", "20240101", "m 1,-1 2,8", "20240229"},
		{"missing month day", "20240401", "20240101", "m 31", "20240531"},
		{"equal dates", "20240126", "20240126", "m 26", "20240226"},
		{"future daily", "20240126", "20240202", "d 30", "20240303"},
		{"future weekly", "20240126", "20240204", "w 7", "20240204"},
		{"future monthly", "20240126", "20240201", "m 1", "20240201"},
		{"leap century gap", "20960301", "20960101", "m 29 2", "21040229"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			now, err := time.Parse(DateFormat, tc.now)
			if err != nil {
				t.Fatal(err)
			}
			now = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.FixedZone("local", -5*60*60))
			got, err := NextDate(now, tc.start, tc.rule)
			if err != nil || got != tc.want {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestNextDateInvalid(t *testing.T) {
	now, _ := time.Parse(DateFormat, "20240126")
	for _, rule := range []string{"", " ", "k 34", "d", "d 0", "d -1", "d 401", "d abc", "d 1 2", "y 1", "w", "w 0", "w 8", "w 1,9", "w abc", "w 1,", "w 1 2", "m", "m 0", "m 32", "m -3", "m 4,40", "m 5 1,13", "m 1 0", "m abc", "m 1,", "m 1 2 3", "m 31 2"} {
		t.Run(rule, func(t *testing.T) {
			if got, err := NextDate(now, "20240126", rule); err == nil {
				t.Fatalf("expected error, got %q", got)
			}
		})
	}
	for _, date := range []string{"", "20240230", "20230229", "2024-01-26", "oops"} {
		if _, err := NextDate(now, date, "d 1"); err == nil {
			t.Errorf("expected error for date %q", date)
		}
	}
}

func TestNextDayHandler(t *testing.T) {
	cases := []struct {
		method, query string
		status        int
		want          string
	}{
		{"GET", "now=20240126&date=20240113&repeat=d+7", 200, "20240127"},
		{"GET", "now=invalid&date=20240113&repeat=d+7", 400, ""},
		{"GET", "now=&date=20240113&repeat=d+7", 400, ""},
		{"GET", "date=20990101&repeat=y", 200, "21000101"},
		{"GET", "now=20240126&date=20240126&repeat=", 400, ""},
		{"POST", "now=20240126&date=20240113&repeat=d+7", 405, ""},
	}
	for _, tc := range cases {
		r := httptest.NewRequest(tc.method, "/api/nextdate?"+tc.query, nil)
		w := httptest.NewRecorder()
		nextDayHandler(w, r)
		if w.Code != tc.status || tc.want != "" && w.Body.String() != tc.want {
			t.Errorf("%s: got %d %q", tc.query, w.Code, w.Body.String())
		}
		if w.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
			t.Errorf("unexpected content type: %s", w.Header().Get("Content-Type"))
		}
	}
}

func TestInit(t *testing.T) {
	Init()
	Init()
	w := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(w, httptest.NewRequest("GET", "/api/nextdate?now=20240126&date=20240126&repeat=y", nil))
	if w.Code != http.StatusOK || w.Body.String() != "20250126" {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}
}
