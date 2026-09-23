package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCardCacheAndQuota(t *testing.T) {
	calls := 0
	fail := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if fail {
			w.WriteHeader(429)
			return
		}
		w.Write([]byte(`{"success":true,"data":[{"id":"bout-1","fighters":[{"fighterName":"One","corner":"red"},{"fighterName":"Two","corner":"blue"}],"winnerFighterSlug":"one"}]}`))
	}))
	defer upstream.Close()
	a := &App{key: "test", base: upstream.URL, client: upstream.Client(), file: filepath.Join(t.TempDir(), "cache.json"), cache: Cache{Events: []Event{{ID: "ufc-test"}}}}
	get := func(id string) map[string]any {
		w := httptest.NewRecorder()
		a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/events/"+id+"/bouts", nil))
		if id == "unknown" {
			if w.Code != 404 {
				t.Fatal(w.Code)
			}
			return nil
		}
		var s map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	get("unknown")
	if calls != 0 {
		t.Fatal("unknown event spent quota")
	}
	get("ufc-test")
	get("ufc-test")
	if calls != 1 || a.cache.Requests != 1 {
		t.Fatal("cache miss")
	}
	bytes, _ := os.ReadFile(a.file)
	var saved Cache
	if err := json.Unmarshal(bytes, &saved); err != nil || len(saved.Cards["ufc-test"].Data) == 0 {
		t.Fatal("card not persisted")
	}
	c := a.cache.Cards["ufc-test"]
	c.Updated = time.Now().Add(-13 * time.Hour)
	c.Attempt = time.Now().Add(-2 * time.Hour)
	a.cache.Cards["ufc-test"] = c
	fail = true
	s := get("ufc-test")
	get("ufc-test")
	if calls != 2 || len(s["bouts"].([]any)) != 1 || s["message"] == "" {
		t.Fatal("stale fallback/backoff failed")
	}
	c = a.cache.Cards["ufc-test"]
	c.Attempt = time.Now().Add(-2 * time.Hour)
	a.cache.Cards["ufc-test"] = c
	a.cache.Requests = 450
	get("ufc-test")
	if calls != 2 {
		t.Fatal("monthly budget exceeded")
	}
	a.cache.Requests = 0
	c.Attempt = time.Now().Add(-2 * time.Hour)
	a.cache.Cards["ufc-test"] = c
	a.cache.RequestTimes = make([]time.Time, 10)
	for i := range a.cache.RequestTimes {
		a.cache.RequestTimes[i] = time.Now()
	}
	get("ufc-test")
	if calls != 2 {
		t.Fatal("minute budget exceeded")
	}
}
