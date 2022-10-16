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

type SlackUrlVerify struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}

type T struct {
	Type     string `json:"type"`
	User     string `json:"user"`
	Reaction string `json:"reaction"`
	Item     struct {
		Type    string `json:"type"`
		Channel string `json:"channel"`
		Ts      string `json:"ts"`
	} `json:"item"`
	ItemUser string `json:"item_user"`
	EventTs  string `json:"event_ts"`
}

type SlackEvent struct {
	ClientMsgId string `json:"client_msg_id"`
	Type        string `json:"type"`
	Text        string `json:"text"`
	User        string `json:"user"`
	Ts          string `json:"ts"`
	Team        string `json:"team"`
	Blocks      []struct {
		Type     string `json:"type"`
		BlockId  string `json:"block_id"`
		Elements []struct {
			Type     string `json:"type"`
			Elements []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"elements"`
		} `json:"elements"`
	} `json:"blocks"`
	Channel     string `json:"channel"`
	EventTs     string `json:"event_ts"`
	ChannelType string `json:"channel_type"`
	Reaction    string `json:"reaction"`
	Item        struct {
		Type    string `json:"type"`
		Channel string `json:"channel"`
		Ts      string `json:"ts"`
	} `json:"item"`
	ItemUser string `json:"item_user"`
}

type SlackCall struct {
	Token          string `json:"token"`
	TeamId         string `json:"team_id"`
	ApiAppId       string `json:"api_app_id"`
	Event          SlackEvent
	Type           string `json:"type"`
	EventId        string `json:"event_id"`
	EventTime      int    `json:"event_time"`
	Authorizations []struct {
		EnterpriseId        interface{} `json:"enterprise_id"`
		TeamId              string      `json:"team_id"`
		UserId              string      `json:"user_id"`
		IsBot               bool        `json:"is_bot"`
		IsEnterpriseInstall bool        `json:"is_enterprise_install"`
	} `json:"authorizations"`
	IsExtSharedChannel bool   `json:"is_ext_shared_channel"`
	EventContext       string `json:"event_context"`
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

	var urlVerify SlackUrlVerify
	err = json.Unmarshal(body, &urlVerify)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	t := urlVerify.Type

	if t == "url_verification" {
		log.Println("received url verification request")
		log.Printf("%v\n", urlVerify)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(urlVerify.Challenge))
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

	var event SlackCall
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("json unmarhal error: %v", err)
	}

	if err := cli.put(ctx, event.Event); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
