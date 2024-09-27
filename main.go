package main

import "net/http"

func healthz(wr http.ResponseWriter, request *http.Request) {
	wr.Header().Add("Content-Type", "text/plain; charset=utf-8")
	wr.Write([]byte("OK"))
}
func main() {
	mux := http.NewServeMux()

	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	mux.Handle("/healthz", http.HandlerFunc(healthz))

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}
