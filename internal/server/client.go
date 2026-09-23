package server

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Called under a.mu. All endpoints share a persisted monthly and rolling minute budget.
func (a *App) request(r *http.Request, path string) ([]byte, error) {
	now := time.Now().UTC()
	if a.cache.Month != now.Format("2006-01") {
		a.cache.Month = now.Format("2006-01")
		a.cache.Requests = 0
	}
	recent := []time.Time{}
	for _, t := range a.cache.RequestTimes {
		if now.Sub(t) < time.Minute {
			recent = append(recent, t)
		}
	}
	if a.cache.Requests >= 450 || len(recent) >= 10 {
		return nil, fmt.Errorf("local API budget exhausted")
	}
	a.cache.Requests++
	a.cache.RequestTimes = append(recent, now)
	if err := a.save(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(r.Context(), "GET", a.base+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", a.key)
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Cito API: HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 4<<20))
}
