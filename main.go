package main

import (
	"cloud.google.com/go/datastore"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"os"
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

type dataStoreClient struct {
	client  *datastore.Client
	taskKey *datastore.Key
}

func newDataStoreClient(ctx context.Context, project string) (dataStoreClient, error) {
	cli, err := datastore.NewClient(ctx, project)
	if err != nil {
		log.Printf("Failed to create datastore client: %v", err)
		return dataStoreClient{}, err
	}
	return dataStoreClient{
		client: cli,
	}, nil
}

func (dc *dataStoreClient) generateKey(kind string, name string) {
	dc.taskKey = datastore.NameKey(kind, name, nil)
}

func (dc *dataStoreClient) put(ctx context.Context, entity SlackEvent) error {
	if _, err := dc.client.Put(ctx, dc.taskKey, &entity); err != nil {
		log.Printf("Failed to save entity: %v\n", err)
		return err
	}
	return nil
}

func (dc *dataStoreClient) close() {
	dc.client.Close()
}

type SlackEvent struct {
	Message string
	User    string
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

	log.Printf("%v\n", string(body))
	ctx := context.Background()
	project := os.Getenv("PROJECT_ID")
	cli, err := newDataStoreClient(ctx, project)
	if err != nil {
		log.Println("Failed initialize app.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer cli.close()

	u, _ := uuid.NewRandom()
	uu := u.String()
	cli.generateKey("slack", uu)

	entity := SlackEvent{
		Message: "hello",
		User:    "Gopher",
	}
	log.Printf("%v\n", entity)
	if err := cli.put(ctx, entity); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
