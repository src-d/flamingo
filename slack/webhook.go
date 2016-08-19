package slack

import (
	"bytes"
	"encoding/json"
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/mvader/slack"
)

// WebhookService is a service to handle slack interactive messages callbacks.
type WebhookService struct {
	token     string
	callbacks chan slack.AttachmentActionCallback
}

// NewWebhookService returns a new WebhookService with the given token.
func NewWebhookService(token string) *WebhookService {
	return &WebhookService{
		token:     token,
		callbacks: make(chan slack.AttachmentActionCallback, 1),
	}
}

// Consume returns a channel where callbacks will be sent.
func (s *WebhookService) Consume() <-chan slack.AttachmentActionCallback {
	return s.callbacks
}

// ServeHTTP is the actual HTTP handler of the service.
func (s *WebhookService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload := r.PostFormValue("payload")
	var callback slack.AttachmentActionCallback
	err := json.NewDecoder(bytes.NewBuffer([]byte(payload))).Decode(&callback)
	if err != nil {
		log15.Error("error decoding request body", "err", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if callback.Token != s.token {
		log15.Warn("received action callback token does not match", "token", callback.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	s.callbacks <- callback
	w.WriteHeader(http.StatusOK)
}
