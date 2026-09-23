package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
	"ufcfans/internal/config"
)

type App struct {
	mu              sync.Mutex
	cache           Cache
	key, base, file string
	client          *http.Client
}

// New creates the server and restores its persisted cache and request budget.
func New(cfg config.Config) (*App, error) {
	a := &App{key: cfg.APIKey, base: "https://api.citoapi.com/api/v1", file: cfg.CacheFile, client: &http.Client{Timeout: 20 * time.Second}}
	if b, err := os.ReadFile(a.file); err == nil {
		if err = json.Unmarshal(b, &a.cache); err != nil {
			return nil, fmt.Errorf("invalid cache: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	return a, nil
}
