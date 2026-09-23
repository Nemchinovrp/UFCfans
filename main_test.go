package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCalendarCacheAndFailureBackoff(t *testing.T) {
	calls := 0
	fail := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("x-api-key") != "secret" {
			t.Error("missing key")
		}
		if fail {
			w.WriteHeader(429)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		status := "next"
		if strings.HasSuffix(r.URL.Path, "recent") {
			status = "past"
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true, "data": []map[string]string{{"slug": status, "name": "UFC test", "date": "2026-09-27", "venue": "Arena"}}})
	}))
	defer upstream.Close()
	a := &App{key: "secret", base: upstream.URL, file: filepath.Join(t.TempDir(), "cache.json"), client: upstream.Client()}
	get := func() {
		w := httptest.NewRecorder()
		a.handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/events", nil))
		if w.Code != 200 {
			t.Fatal(w.Code)
		}
		if strings.Contains(w.Body.String(), "secret") {
			t.Fatal("key exposed")
		}
	}
	get()
	get()
	if calls != 2 || len(a.cache.Events) != 2 {
		t.Fatalf("calls=%d events=%d", calls, len(a.cache.Events))
	}
	var persisted Cache
	b, err := os.ReadFile(a.file)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &persisted); err != nil || persisted.Requests != 2 {
		t.Fatal("quota not persisted")
	}
	a.cache.Updated = time.Now().Add(-13 * time.Hour)
	a.cache.Attempt = time.Now().Add(-2 * time.Hour)
	fail = true
	get()
	get()
	if calls != 3 || len(a.cache.Events) != 2 {
		t.Fatal("stale data lost or retry not throttled")
	}
	a.cache.Attempt = time.Now().Add(-2 * time.Hour)
	a.cache.Requests = 449
	get()
	if calls != 3 {
		t.Fatal("budget exceeded")
	}
}
func TestNoKeyMakesNoRequests(t *testing.T) {
	a := &App{}
	w := httptest.NewRecorder()
	a.handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/events", nil))
	if !strings.Contains(w.Body.String(), `"configured":false`) {
		t.Fatal(w.Body.String())
	}
}
func TestDecode(t *testing.T) {
	for _, body := range []string{`{"success":true,"data":[{"slug":"ufc","name":"UFC","date":"2026-09-27"}]}`, `{"data":{"events":[{"slug":"ufc","date":"2026-09-27T22:00:00Z"}]}}`} {
		events, err := decodeEvents(strings.NewReader(body), "upcoming")
		if err != nil || len(events) != 1 {
			t.Fatalf("%v %v", events, err)
		}
	}
	for _, body := range []string{`{"success":false,"data":[]}`, `{"data":{}}`, `{"data":[{"date":"nonsense"}]}`} {
		if _, err := decodeEvents(strings.NewReader(body), "upcoming"); err == nil {
			t.Fatal("expected error")
		}
	}
}
