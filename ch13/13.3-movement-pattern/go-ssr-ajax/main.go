package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)


type State struct {
Now string `json:"now"`
Items []string `json:"items"`
}


func state() State {
return State{Now: time.Now().Format(time.RFC3339), Items: []string{"apple","banana","cherry"}}
}


func handleIndex(w http.ResponseWriter, r *http.Request) {
t := template.Must(template.ParseFiles(filepath.Join("templates", "index.html")))
_ = t.Execute(w, state())
}


func handleAPI(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(state())
}


func main() {
http.HandleFunc("/", handleIndex)
http.HandleFunc("/api/state", handleAPI)
log.Println("Go SSR+Ajax at :8082")
log.Fatal(http.ListenAndServe(":8082", nil))
}