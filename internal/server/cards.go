package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type CardCache struct {
	Data    json.RawMessage `json:"data,omitempty"`
	Updated time.Time       `json:"updatedAt"`
	Attempt time.Time       `json:"attemptAt"`
}

func (a *App) card(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := r.PathValue("id")
	known := false
	for _, e := range a.cache.Events {
		if e.ID == id {
			known = true
			break
		}
	}
	if !known {
		http.Error(w, "Турнир не найден", 404)
		return
	}
	if a.cache.Cards == nil {
		a.cache.Cards = map[string]CardCache{}
	}
	now := time.Now().UTC()
	c := a.cache.Cards[id]
	message := ""
	stale := c.Updated.IsZero() || now.Sub(c.Updated) >= 12*time.Hour
	if stale {
		switch {
		case a.key == "":
			message = "Для загрузки боёв нужен ключ Cito API."
		case now.Sub(c.Attempt) < time.Hour:
			message = "Обновление временно недоступно. Повтор возможен через час после последней попытки."
		default:
			c.Attempt = now
			a.cache.Cards[id] = c
			body, err := a.request(r, "/ufc/events/"+url.PathEscape(id)+"/bouts")
			if err == nil {
				var envelope struct {
					Success *bool           `json:"success"`
					Data    json.RawMessage `json:"data"`
				}
				err = json.Unmarshal(body, &envelope)
				if err == nil && envelope.Success != nil && !*envelope.Success {
					err = fmt.Errorf("API success=false")
				}
				var rows []json.RawMessage
				if err == nil {
					err = json.Unmarshal(envelope.Data, &rows)
					if err != nil {
						var nested struct {
							Bouts []json.RawMessage `json:"bouts"`
						}
						err = json.Unmarshal(envelope.Data, &nested)
						rows = nested.Bouts
					}
					if rows == nil {
						err = fmt.Errorf("missing bouts array")
					}
				}
				if err == nil {
					c.Data, _ = json.Marshal(rows)
					c.Updated = now
					a.cache.Cards[id] = c
					err = a.save()
				}
			}
			if err != nil {
				log.Printf("card sync: %v", err)
				message = "Не удалось обновить кард. Проверьте доступность API и лимит запросов; повтор — через час."
			}
		}
	}
	data := c.Data
	if len(data) == 0 {
		data = json.RawMessage("[]")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]any{"bouts": data, "updatedAt": c.Updated, "stale": c.Updated.IsZero() || now.Sub(c.Updated) >= 12*time.Hour, "message": message, "requests": a.cache.Requests})
}
