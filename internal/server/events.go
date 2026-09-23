package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

type Event struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Date     string `json:"date"`
	Venue    string `json:"venue"`
	Location string `json:"location"`
	Status   string `json:"status"`
}

func str(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s, ok := m[k].(string); ok && s != "" {
			return s
		}
	}
	return ""
}
func decodeEvents(r io.Reader, status string) ([]Event, error) {
	var envelope struct {
		Success *bool           `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(r).Decode(&envelope); err != nil {
		return nil, err
	}
	if envelope.Success != nil && !*envelope.Success {
		return nil, fmt.Errorf("API success=false")
	}
	var rows []map[string]any
	if err := json.Unmarshal(envelope.Data, &rows); err != nil {
		var nested struct {
			Events []map[string]any `json:"events"`
		}
		if err = json.Unmarshal(envelope.Data, &nested); err != nil || nested.Events == nil {
			return nil, fmt.Errorf("unexpected events response")
		}
		rows = nested.Events
	}
	if rows == nil {
		return nil, fmt.Errorf("missing event data")
	}
	events := make([]Event, 0, len(rows))
	for _, m := range rows {
		e := Event{ID: str(m, "slug", "dataId", "id"), Name: str(m, "name", "title", "eventName"), Date: str(m, "startDate", "date", "eventDate", "startTime", "startsAt", "dateTime"), Venue: str(m, "venue", "venueName", "arena"), Location: str(m, "location", "city"), Status: status}
		if v, ok := m["venue"].(map[string]any); ok {
			e.Venue = str(v, "name")
			if e.Location == "" {
				e.Location = str(v, "city", "location")
			}
		}
		if e.Name == "" {
			e.Name = strings.ReplaceAll(e.ID, "-", " ")
		}
		if e.ID == "" {
			e.ID = e.Name + e.Date
		}
		if e.Date != "" {
			valid := false
			for _, layout := range []string{time.RFC3339, "2006-01-02", "2006-01-02 15:04:05"} {
				if t, err := time.Parse(layout, e.Date); err == nil {
					if layout != "2006-01-02" {
						e.Date = t.Format(time.RFC3339)
					}
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("unsupported event date")
			}
		}
		events = append(events, e)
	}
	return events, nil
}
func (a *App) fetch(ctx context.Context, path, status string) ([]Event, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost", nil)
	body, err := a.request(req, path)
	if err != nil {
		return nil, err
	}
	return decodeEvents(bytes.NewReader(body), status)
}
func (a *App) calendar(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now().UTC()
	message := ""
	configured := a.key != ""
	if a.cache.Month != now.Format("2006-01") {
		a.cache.Month = now.Format("2006-01")
		a.cache.Requests = 0
	}
	stale := a.cache.Updated.IsZero() || now.Sub(a.cache.Updated) >= 12*time.Hour
	if configured && stale && now.Sub(a.cache.Attempt) >= time.Hour {
		if a.cache.Requests > 448 {
			message = "Локальный лимит запросов исчерпан. Следующее обновление — в новом месяце."
		} else {
			a.cache.Attempt = now
			upcoming, err := a.fetch(r.Context(), "/ufc/events/upcoming", "upcoming")
			if err == nil {
				var recent []Event
				recent, err = a.fetch(r.Context(), "/ufc/events/recent", "completed")
				if err == nil {
					byID := map[string]Event{}
					for _, e := range a.cache.Events {
						byID[e.ID] = e
					}
					for _, e := range append(upcoming, recent...) {
						byID[e.ID] = e
					}
					merged := make([]Event, 0, len(byID))
					for _, e := range byID {
						merged = append(merged, e)
					}
					sort.Slice(merged, func(i, j int) bool { return merged[i].Date < merged[j].Date })
					a.cache.Events = merged
					a.cache.Updated = now
					if err = a.save(); err != nil {
						message = "Данные получены, но сохранить кеш на диск не удалось."
						log.Print(err)
					}
				}
			}
			if err != nil {
				log.Printf("calendar sync: %v", err)
				message = "Не удалось обновить данные Cito API. Проверьте ключ, лимит и доступность сервиса. Повтор — не раньше чем через час."
			}
		}
	}
	stale = a.cache.Updated.IsZero() || now.Sub(a.cache.Updated) >= 12*time.Hour
	if configured && stale && message == "" {
		message = "Показан сохранённый календарь. Обновление временно недоступно; повторные запросы ограничены для экономии тарифа."
	}
	events := a.cache.Events
	if events == nil {
		events = []Event{}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]any{"events": events, "configured": configured, "updatedAt": a.cache.Updated, "stale": stale, "requests": a.cache.Requests, "budget": 450, "message": message})
}
