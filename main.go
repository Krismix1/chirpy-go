package main

import (
	"net/http"
)

func main() {
	apiCfg := apiConfig{}
	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /metrics", apiCfg.metrics)
	mux.HandleFunc("POST /reset", apiCfg.reset)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}
