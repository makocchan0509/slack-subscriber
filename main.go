package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)

func main() {
	http.HandleFunc("/slack/events", slackNotification)
	http.HandleFunc("/health", healthCheck)
	port := ":8080"
	log.Printf("Start Server on %s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func slackNotification(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := make([]byte, length)
	length, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var jsonBody map[string]interface{}
	err = json.Unmarshal(body[:length], &jsonBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if v, ok := jsonBody["type"]; ok && v.(string) == "url_verification" {
		log.Println("received url verification request")
		log.Printf("%v\n", jsonBody)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonBody["challenge"].(string)))

		return
	}
	log.Println("TODO logic")
	log.Printf("%v\n", jsonBody)
	w.WriteHeader(http.StatusOK)
	return
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
