package main

import (
  "fmt"
  "log"
  "net/http"
)

func homePage(w http.ResponseWritter, r *http.Request) {
  fmt.Fprintf(w, "Homepage Endpoint hit")
}

func handleRequests() {
  http.handleFunc("/", homePage)
  log.Fatal(http.ListenAndServe(":8081", nil))
}

func main() {
  handleRequests()
}
