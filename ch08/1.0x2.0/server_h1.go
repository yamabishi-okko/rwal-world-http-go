package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func slow(w http.ResponseWriter, r *http.Request) {
	// ?d=500 のように指定したミリ秒だけ待つ
	dms := r.URL.Query().Get("d")
	ms, _ := strconv.Atoi(dms)
	if ms <= 0 {
		ms = 500
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	fmt.Fprintf(w, "h1 slow done: %dms\n", ms)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", slow)

	addr := ":8081"
	log.Println("HTTP/1.1 server on", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
