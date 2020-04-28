package main

import (
	"fmt"
	"log"
	"net/http"
	"html"
)

func main() {
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})
	http.HandleFunc("/install", func(w http.ResponseWriter, r *http.Request) {
		// Get the query param
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})
	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		// Get the query param
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})
	log.Println("Starting up...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
