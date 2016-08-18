package slack

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testCallback = `{
  "actions": [
    {
      "name": "recommend",
      "value": "yes"
    }
  ],
  "callback_id": "bot::channel::test_callback",
  "team": {
    "id": "T47563693",
    "domain": "watermelonsugar"
  },
  "channel": {
    "id": "C065W1189",
    "name": "forgotten-works"
  },
  "user": {
    "id": "U045VRZFT",
    "name": "brautigan"
  },
  "action_ts": "1458170917.164398",
  "message_ts": "1458170866.000004",
  "attachment_id": "1",
  "token": "xAB3yVzGS4BQ3O9FACTa8Ho4",
  "original_message": "{\"text\":\"New comic book alert!\",\"attachments\":[{\"title\":\"The Further Adventures of Slackbot\",\"fields\":[{\"title\":\"Volume\",\"value\":\"1\",\"short\":true},{\"title\":\"Issue\",\"value\":\"3\",\"short\":true}],\"author_name\":\"Stanford S. Strickland\",\"author_icon\":\"https://api.slack.com/img/api/homepage_custom_integrations-2x.png\",\"image_url\":\"http://i.imgur.com/OJkaVOI.jpg?1\"},{\"title\":\"Synopsis\",\"text\":\"After @episod pushed exciting changes to a devious new branch back in Issue 1, Slackbot notifies @don about an unexpected deploy...\"},{\"fallback\":\"Would you recommend it to customers?\",\"title\":\"Would you recommend it to customers?\",\"callback_id\":\"comic_1234_xyz\",\"color\":\"#3AA3E3\",\"attachment_type\":\"default\",\"actions\":[{\"name\":\"recommend\",\"text\":\"Recommend\",\"type\":\"button\",\"value\":\"recommend\"},{\"name\":\"no\",\"text\":\"No\",\"type\":\"button\",\"value\":\"bad\"}]}]}",
  "response_url": "https://hooks.slack.com/actions/T47563693/6204672533/x7ZLaiVMoECAW50Gw1ZYAXEM"
}`

func TestWebhook(t *testing.T) {
	assert := assert.New(t)
	cases := []struct {
		body          string
		status        int
		shouldConsume bool
		callback      string
	}{
		{"", http.StatusBadRequest, false, ""},
		{"skdjadljsal", http.StatusBadRequest, false, ""},
		{`{"token": "fooo"}`, http.StatusUnauthorized, false, ""},
		{testCallback, http.StatusOK, true, "bot::channel::test_callback"},
	}

	for _, c := range cases {
		w := NewWebhookService("xAB3yVzGS4BQ3O9FACTa8Ho4")
		srv := httptest.NewServer(w)
		data := url.Values{}
		data.Set("payload", c.body)

		resp, err := http.Post(srv.URL, "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
		assert.Nil(err)
		assert.Equal(resp.StatusCode, c.status)

		if c.shouldConsume {
			select {
			case cb := <-w.Consume():
				assert.Equal(cb.CallbackID, c.callback)
			case <-time.After(10 * time.Millisecond):
				assert.FailNow("timeout")
			}
		}
		srv.Close()
	}
}
