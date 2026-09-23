package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"time"
)

func (a *App) knownFighter(slug string) bool {
	for _, card := range a.cache.Cards {
		var bouts []struct {
			Fighters []struct {
				Slug string `json:"fighterSlug"`
			} `json:"fighters"`
		}
		if json.Unmarshal(card.Data, &bouts) != nil {
			continue
		}
		for _, bout := range bouts {
			for _, fighter := range bout.Fighters {
				if fighter.Slug == slug {
					return true
				}
			}
		}
	}
	return false
}
func (a *App) fighter(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	slug := r.PathValue("slug")
	if !a.knownFighter(slug) {
		http.Error(w, "Боец не найден в загруженных кардах", 404)
		return
	}
	if a.cache.Fighters == nil {
		a.cache.Fighters = map[string]CardCache{}
	}
	c := a.cache.Fighters[slug]
	now := time.Now().UTC()
	message := ""
	if c.Updated.IsZero() || now.Sub(c.Updated) >= 24*time.Hour {
		switch {
		case a.key == "":
			message = "Для загрузки профиля нужен ключ Cito API."
		case now.Sub(c.Attempt) < time.Hour:
			message = "Повторное обновление будет доступно через час после последней попытки."
		default:
			c.Attempt = now
			a.cache.Fighters[slug] = c
			body, err := a.request(r, "/ufc/fighters/"+url.PathEscape(slug))
			if err == nil {
				var envelope struct {
					Success *bool           `json:"success"`
					Data    json.RawMessage `json:"data"`
				}
				err = json.Unmarshal(body, &envelope)
				if err == nil && envelope.Success != nil && !*envelope.Success {
					err = fmt.Errorf("API success=false")
				}
				var profile struct {
					Name string `json:"name"`
					Slug string `json:"slug"`
				}
				if err == nil {
					err = json.Unmarshal(envelope.Data, &profile)
					if profile.Name == "" || profile.Slug != slug {
						err = fmt.Errorf("invalid fighter profile")
					}
				}
				if err == nil {
					c.Data = envelope.Data
					c.Updated = now
					a.cache.Fighters[slug] = c
					err = a.save()
				}
			}
			if err != nil {
				log.Printf("fighter sync: %v", err)
				message = "Не удалось обновить профиль. Проверьте доступность API и лимит. Повтор — через час."
			}
		}
	}
	data := c.Data
	if len(data) == 0 {
		data = json.RawMessage("null")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	var normalized any
	json.Unmarshal(data, &normalized)
	json.NewEncoder(w).Encode(map[string]any{"fighter": normalizeDecimals(normalized), "updatedAt": c.Updated, "stale": c.Updated.IsZero() || now.Sub(c.Updated) >= 24*time.Hour, "message": message})
}

// Cito sometimes serializes Decimal.js numbers as sign/exponent/base-1e7 digits.
func normalizeDecimals(value any) any {
	switch v := value.(type) {
	case map[string]any:
		sign, sok := v["s"].(float64)
		exponent, eok := v["e"].(float64)
		digits, dok := v["d"].([]any)
		if len(v) == 3 && sok && eok && dok && len(digits) > 0 && (sign == 1 || sign == -1) {
			first, ok := digits[0].(float64)
			if ok && first >= 0 && first < 1e7 {
				n := first
				valid := true
				for i, d := range digits[1:] {
					part, ok := d.(float64)
					if !ok || part < 0 || part >= 1e7 {
						valid = false
						break
					}
					n += part * math.Pow(1e7, -float64(i+1))
				}
				width := 1.0
				if first > 0 {
					width = math.Floor(math.Log10(first)) + 1
				}
				result := sign * n * math.Pow10(int(exponent-width+1))
				if valid && !math.IsInf(result, 0) && !math.IsNaN(result) {
					return result
				}
			}
		}
		for k, child := range v {
			v[k] = normalizeDecimals(child)
		}
		return v
	case []any:
		for i, child := range v {
			v[i] = normalizeDecimals(child)
		}
		return v
	default:
		return value
	}
}
