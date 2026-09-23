package server

import (
	"net/http"
	"path/filepath"
	"strings"
	"ufcfans/web"
)

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/events", a.calendar)
	mux.HandleFunc("GET /api/fighters/{slug}", a.fighter)
	mux.HandleFunc("GET /api/events/{id}/bouts", a.card)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		if path != "/index.html" && path != "/style.css" && path != "/app.js" {
			http.NotFound(w, r)
			return
		}
		b, err := web.Assets.ReadFile(strings.TrimPrefix(path, "/"))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		types := map[string]string{".html": "text/html; charset=utf-8", ".css": "text/css; charset=utf-8", ".js": "text/javascript; charset=utf-8"}
		w.Header().Set("Content-Type", types[filepath.Ext(path)])
		w.Write(b)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self'; img-src 'self' data: https://ufc.com https://www.ufc.com https://api.citoapi.com; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
		mux.ServeHTTP(w, r)
	})
}
