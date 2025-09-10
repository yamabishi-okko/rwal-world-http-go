package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func slow(w http.ResponseWriter, r *http.Request) {
	dms := r.URL.Query().Get("d")
	ms, _ := strconv.Atoi(dms)
	if ms <= 0 {
		ms = 500
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	fmt.Fprintf(w, "h2 slow done: %dms\n", ms)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", slow)

	h2s := &http2.Server{}
	handler := h2c.NewHandler(mux, h2s)

	addr := ":8082"
	log.Println("HTTP/2 (h2c) server on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
