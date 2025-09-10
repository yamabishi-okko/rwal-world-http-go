package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	log.Println("WebTransport demo => http://localhost:18888")
	log.Fatal(http.ListenAndServe(":18888", nil))
}
