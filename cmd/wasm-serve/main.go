package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusNoContent)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")

		http.ServeFile(w, r, "ui/index.html")
	})
	http.HandleFunc("/tileset.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "image/png")

		http.ServeFile(w, r, "tilesets/tileset.png")
	})
	http.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "application/javascript")

		http.ServeFile(w, r, "ui/wasm_exec.js")
	})
	http.HandleFunc("/hashira.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "application/javascript")

		http.ServeFile(w, r, "ui/hashira.js")
	})
	http.HandleFunc("/hashira.wasm", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Type", "application/wasm")

		http.ServeFile(w, r, "bin/hashira.wasm")
	})
	http.ListenAndServe(":3000", nil)
}
