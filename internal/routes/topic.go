package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"push_article/pkg/token"

	"firebase.google.com/go/v4/messaging"
	"github.com/go-chi/chi"
)

type TopicService struct {
	token.Storage
	*messaging.Client
}

func (s *TopicService) subscribe(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request struct {
		UserID uint64 `json:"user_id"`
		Topic  string `json:"topic"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}

	tokens, err := s.UserTokens(r.Context(), request.UserID)
	if err != nil {
		http.Error(w, "user tokens get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(tokens) == 0 {
		http.Error(w, "user has no tokens", http.StatusBadRequest)
		return
	}

	log.Printf("Subscribe tokens %v to topic %s", tokens, request.Topic)

	resp, err := s.SubscribeToTopic(r.Context(), tokens.Values(), request.Topic)
	if err != nil {
		http.Error(w, "token subscribe error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(resp.Errors) > 0 {
		log.Printf("Subscribe errors: %v", resp.Errors)
	}
}

func (s *TopicService) unsubscribe(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request struct {
		UserID uint64 `json:"user_id"`
		Topic  string `json:"topic"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}

	tokens, err := s.UserTokens(r.Context(), request.UserID)
	if err != nil {
		http.Error(w, "user tokens get failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(tokens) == 0 {
		http.Error(w, "user has no tokens", http.StatusBadRequest)
	}

	log.Printf("Unsubscribe tokens %v to topic %s", tokens, request.Topic)

	resp, err := s.UnsubscribeFromTopic(r.Context(), tokens.Values(), request.Topic)
	if err != nil {
		http.Error(w, "token unsubscribe error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(resp.Errors) > 0 {
		log.Printf("Unsubscribe errors: %v", resp.Errors)
	}
}

func (s *TopicService) AddToRouter(r chi.Router) {
	r.Post("/subscribe", s.subscribe)
	r.Post("/unsubscribe", s.unsubscribe)
}
