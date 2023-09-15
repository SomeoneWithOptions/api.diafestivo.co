package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Message struct {
    Text string `json:"message"`
}

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        message := Message{Text: "hello"}
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(message)
				current_time := time.Now()
				fmt.Println(current_time)
    })

    http.ListenAndServe(":3002", nil)
}
