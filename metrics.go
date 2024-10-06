package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (ac *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ac.fileServerHits.Add(1)
		next.ServeHTTP(rw, req)
	})
}

func (ac *apiConfig) metrics(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/html; charset=utf-8")
	rw.Write([]byte(fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, ac.fileServerHits.Load())))
}

func (ac *apiConfig) reset(rw http.ResponseWriter, req *http.Request) {
	ac.fileServerHits.Store(0)
}
