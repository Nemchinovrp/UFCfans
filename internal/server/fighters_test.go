package server

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestDecimalNormalization(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  float64
	}{{`{"s":1,"e":-1,"d":[5800000]}`, 0.58}, {`{"s":1,"e":2,"d":[124,5000000]}`, 124.5}, {`{"s":1,"e":0,"d":[8,2600000]}`, 8.26}, {`{"s":1,"e":1,"d":[65]}`, 65}, {`{"s":1,"e":0,"d":[0]}`, 0}} {
		var v any
		json.Unmarshal([]byte(tc.input), &v)
		got, ok := normalizeDecimals(v).(float64)
		if !ok || math.Abs(got-tc.want) > 1e-8 {
			t.Fatalf("got %v want %v", got, tc.want)
		}
	}
}
func TestFighterCache(t *testing.T) {
	calls := 0
	invalid := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if invalid {
			w.Write([]byte(`{"data":{"slug":"wrong","name":"Wrong"}}`))
			return
		}
		w.Write([]byte(`{"data":{"slug":"test-fighter","name":"Test Fighter","stats":{"strikingAccuracy":{"s":1,"e":-1,"d":[5800000]}}}}`))
	}))
	defer upstream.Close()
	a := &App{key: "test", base: upstream.URL, client: upstream.Client(), file: filepath.Join(t.TempDir(), "cache.json"), cache: Cache{Cards: map[string]CardCache{"event": {Data: json.RawMessage(`[{"fighters":[{"fighterSlug":"test-fighter"}]}]`)}}}}
	get := func(slug string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		a.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/fighters/"+slug, nil))
		return w
	}
	if get("unknown").Code != 404 || calls != 0 {
		t.Fatal("unknown fighter consumed quota")
	}
	w := get("test-fighter")
	get("test-fighter")
	if calls != 1 {
		t.Fatal("cache missed")
	}
	var body struct {
		Fighter struct {
			Stats struct {
				Accuracy float64 `json:"strikingAccuracy"`
			} `json:"stats"`
		} `json:"fighter"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Fighter.Stats.Accuracy != 0.58 {
		t.Fatal("decimal conversion failed", err)
	}
	c := a.cache.Fighters["test-fighter"]
	c.Updated = time.Now().Add(-25 * time.Hour)
	c.Attempt = time.Now().Add(-2 * time.Hour)
	a.cache.Fighters["test-fighter"] = c
	invalid = true
	get("test-fighter")
	get("test-fighter")
	if calls != 2 || string(a.cache.Fighters["test-fighter"].Data) != string(c.Data) {
		t.Fatal("invalid response replaced cache or backoff failed")
	}
	c = a.cache.Fighters["test-fighter"]
	c.Attempt = time.Now().Add(-2 * time.Hour)
	a.cache.Fighters["test-fighter"] = c
	a.cache.Requests = 450
	get("test-fighter")
	if calls != 2 {
		t.Fatal("quota exceeded")
	}
}
