package slack

import (
	"encoding/json"
	"io/ioutil"
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
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var callback slack.AttachmentActionCallback
	if err := json.Unmarshal(bytes, &callback); err != nil {
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
