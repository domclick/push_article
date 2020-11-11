package routes

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"push_article/pkg/token"

	"firebase.google.com/go/v4/messaging"
	"github.com/go-chi/chi"
)

type NotificationService struct {
	*messaging.Client
	token.Storage
}

func (ns *NotificationService) sendNotification(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request struct {
		UserID       uint64                   `json:"user_id"`
		Data         map[string]string        `json:"data"`
		Notification *messaging.Notification  `json:"notification"`
		Android      *messaging.AndroidConfig `json:"android"`
		Webpush      *messaging.WebpushConfig `json:"webpush"`
		APNS         *messaging.APNSConfig    `json:"apns"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}

	tokens, err := ns.UserTokens(r.Context(), request.UserID)
	if err != nil {
		http.Error(w, "user tokens get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := ns.SendMulticast(r.Context(), &messaging.MulticastMessage{
		Tokens:       tokens.Values(),
		Data:         request.Data,
		Notification: request.Notification,
		Android:      request.Android,
		Webpush:      request.Webpush,
		APNS:         request.APNS,
	})
	if err != nil {
		http.Error(w, "send notifications failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		// we should clean up unregistered tokens
		var unregisteredTokens []string
		for i, item := range resp.Responses {
			if messaging.IsUnregistered(item.Error) {
				unregisteredTokens = append(unregisteredTokens, tokens[i].Token)
			}
		}

		log.Printf("Delete unregistered tokens: %+v", unregisteredTokens)

		err := ns.DeleteTokens(context.Background(), unregisteredTokens)
		if err != nil {
			log.Printf("Delete unregistered tokens failed: %v", err)
		}
	}()
}

func (ns *NotificationService) sendNotificationToTopic(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request struct {
		Topic        string                   `json:"topic"`
		Condition    string                   `json:"condition"`
		Data         map[string]string        `json:"data"`
		Notification *messaging.Notification  `json:"notification"`
		Android      *messaging.AndroidConfig `json:"android"`
		Webpush      *messaging.WebpushConfig `json:"webpush"`
		APNS         *messaging.APNSConfig    `json:"apns"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err = ns.Send(r.Context(), &messaging.Message{
		Topic:        request.Topic,
		Data:         request.Data,
		Notification: request.Notification,
		Android:      request.Android,
		Webpush:      request.Webpush,
		APNS:         request.APNS,
	})
	if err != nil {
		http.Error(w, "send notifications failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ns *NotificationService) confirm(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	log.Printf("Push %s confirmed", chi.URLParam(r, "id"))
}

func (ns *NotificationService) AddToRouter(r chi.Router) {
	r.Post("/", ns.sendNotification)
	r.Post("/topic", ns.sendNotificationToTopic)
	r.Post("/{id}/confirm", ns.confirm)
}
