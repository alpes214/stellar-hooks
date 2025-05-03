package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	log.Printf("Received Webhook: %s", body)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Listening on :9090 for webhook events...")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
