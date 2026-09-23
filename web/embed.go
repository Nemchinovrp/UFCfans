// Package web embeds the browser assets in the application binary.
package web

import "embed"

// Assets contains only public frontend files, never Go sources or configuration.
//
//go:embed index.html style.css app.js
var Assets embed.FS
