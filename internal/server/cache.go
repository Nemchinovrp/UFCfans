package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Cache struct {
	Fighters     map[string]CardCache `json:"fighters,omitempty"`
	Cards        map[string]CardCache `json:"cards,omitempty"`
	RequestTimes []time.Time          `json:"requestTimes,omitempty"`
	Events       []Event              `json:"events"`
	Updated      time.Time            `json:"updatedAt"`
	Attempt      time.Time            `json:"attemptAt"`
	Month        string               `json:"month"`
	Requests     int                  `json:"requests"`
}

func (a *App) save() error {
	if err := os.MkdirAll(filepath.Dir(a.file), 0700); err != nil {
		return err
	}
	b, err := json.Marshal(a.cache)
	if err != nil {
		return err
	}
	if err = os.WriteFile(a.file+".tmp", b, 0600); err != nil {
		return err
	}
	return os.Rename(a.file+".tmp", a.file)
}
