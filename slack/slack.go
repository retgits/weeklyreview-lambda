package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Service is a struct that represents a Slack connection
type Service struct {
	WebhookURL string
}

// Message is a struct that represents a message that is sent to Slack using the Webhook API
type Message struct {
	Channel     string       `json:"channel"`
	Username    string       `json:"username"`
	IconEmoji   string       `json:"icon_emoji"`
	Attachments []Attachment `json:"attachments"`
}

// Attachment represents a single attachment sent with a Slack message
type Attachment struct {
	Fallback   string `json:"fallback"`
	Color      string `json:"color"`
	Pretext    string `json:"pretext"`
	AuthorName string `json:"author_name"`
	Title      string `json:"title"`
	TitleLink  string `json:"title_link"`
	Text       string `json:"text"`
	TimeStamp  int64  `json:"ts"`
}

// Marshal turns the SlackMessage into a byte-array payload so it can be sent to Slack
func (r *Message) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// New creates a new instance of the service
func New() *Service {
	return &Service{}
}

// WithWebhook allows the URL to be set
func (s *Service) WithWebhook(webhook string) *Service {
	s.WebhookURL = webhook
	return s
}

// Send sends the message to Slack using the Webhook API
func (s *Service) Send(m *Message) error {
	payload, err := m.Marshal()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", s.WebhookURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("slack responded with statuscode %d: %s", res.StatusCode, string(body))
	}

	return nil
}
