package slack

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	WebhookSuccessURL = "https://my.slack.localhost/success"
	WebhookFailureURL = "https://my.slack.localhost/fail"
)

func TestService(t *testing.T) {
	assert := assert.New(t)

	svc := New()
	assert.NotNil(svc)

	svc = svc.WithWebhook(WebhookSuccessURL)
	assert.NotNil(svc)
	assert.Equal(svc.WebhookURL, WebhookSuccessURL)

	svc = New().WithWebhook(WebhookSuccessURL)
	assert.NotNil(svc)
	assert.Equal(svc.WebhookURL, WebhookSuccessURL)
}

func TestMarshal(t *testing.T) {
	assert := assert.New(t)

	msg := &Message{}
	bytes, err := msg.Marshal()
	assert.NoError(err)
	assert.NotEmpty(bytes)
}

func TestSend(t *testing.T) {
	assert := assert.New(t)

	msg := &Message{}
	bytes, err := msg.Marshal()
	assert.NoError(err)
	assert.NotEmpty(bytes)

	svc := New().WithWebhook(WebhookSuccessURL)
	assert.NotNil(svc)
	assert.Equal(svc.WebhookURL, WebhookSuccessURL)

	err = svc.Send(msg)
	assert.Error(err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", WebhookSuccessURL, httpmock.NewStringResponder(200, `success...`))
	httpmock.RegisterResponder("POST", WebhookFailureURL, httpmock.NewStringResponder(400, "some arbitrary error..."))

	svc.WebhookURL = WebhookFailureURL
	err = svc.Send(msg)
	assert.Error(err)

	svc.WebhookURL = WebhookSuccessURL
	err = svc.Send(msg)
	assert.NoError(err)
}
