package slack

import (
	"encoding/json"
	"net/http"

	"github.com/nlopes/slack"
)

type WebhookService struct {
	Token     string
	callbacks chan slack.AttachmentActionCallback
}

func NewWebhookService(token string) *WebhookService {
	return &WebhookService{
		Token:     token,
		callbacks: make(chan slack.AttachmentActionCallback, 1),
	}
}

func (s *WebhookService) Consume() <-chan slack.AttachmentActionCallback {
	return s.callbacks
}

func (s *WebhookService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var callback slack.AttachmentActionCallback
	err := json.NewDecoder(r.Body).Decode(&callback)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if callback.Token != s.Token {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	s.callbacks <- callback
	w.WriteHeader(http.StatusOK)
}