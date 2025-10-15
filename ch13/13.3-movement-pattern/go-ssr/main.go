package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)


type ViewData struct {
Now string
Items []string
}


func handleIndex(w http.ResponseWriter, r *http.Request) {
t := template.Must(template.ParseFiles(filepath.Join("templates", "index.html")))
data := ViewData{
Now: time.Now().Format(time.RFC3339),
Items: []string{"apple", "banana", "cherry"},
}
if err := t.Execute(w, data); err != nil { http.Error(w, err.Error(), 500) }
}


func main() {
http.HandleFunc("/", handleIndex)
log.Println("Go SSR at :8081")
log.Fatal(http.ListenAndServe(":8081", nil))
}