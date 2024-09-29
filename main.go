package main

import (
	"net/http"
)

func main() {
	apiCfg := apiConfig{}
	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/metrics", apiCfg.metrics)
	mux.HandleFunc("/reset", apiCfg.reset)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}
