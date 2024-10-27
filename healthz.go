package main

import "net/http"

func handlerHealthz(rw http.ResponseWriter, request *http.Request) {
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.Write([]byte("OK"))
}
