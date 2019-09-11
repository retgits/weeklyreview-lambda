package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/stretchr/testify/assert"
)

const (
	AWSRegion        = "us-west-2"
	TrelloAppKey     = ""
	TrelloAppToken   = ""
	TrelloMemberName = ""
	TrelloBoard      = ""
	TrelloList       = ""
	SlackWebhookURL  = ""
	SlackChannel     = "general"
	SlackUser        = "trello-bot"
	SlackEmoji       = "ghost"
)

func TestHandler(t *testing.T) {
	assert := assert.New(t)

	bytes, err := ioutil.ReadFile("./test/event.json")
	assert.NoError(err)

	var request events.CloudWatchEvent
	err = json.Unmarshal(bytes, &request)
	assert.NoError(err)

	err = handler(request)
	assert.Error(err)

	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(AWSRegion),
	}))

	kmsSvc := kms.New(awsSession)

	os.Setenv("AWS_REGION", AWSRegion)
	os.Setenv("SLACK_CHANNEL", SlackChannel)
	os.Setenv("SLACK_EMOJI", SlackEmoji)
	os.Setenv("SLACK_USER", SlackUser)
	os.Setenv("TRELLO_BOARD", TrelloBoard)
	os.Setenv("TRELLO_LIST", TrelloList)
	os.Setenv("TRELLO_APP_KEY", encodeString(kmsSvc, TrelloAppKey))
	os.Setenv("TRELLO_APP_TOKEN", encodeString(kmsSvc, TrelloAppToken))
	os.Setenv("TRELLO_MEMBER_NAME", encodeString(kmsSvc, TrelloMemberName))
	os.Setenv("SLACK_WEBHOOK_URL", encodeString(kmsSvc, SlackWebhookURL))

	err = handler(request)
	assert.NoError(err)
}

func encodeString(kmsSvc *kms.KMS, payload string) string {
	output, err := kmsSvc.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String("alias/retgits/lambda"),
		Plaintext: []byte(payload),
	})
	if err != nil {
		panic(err.Error())
	}

	return base64.StdEncoding.EncodeToString(output.CiphertextBlob)
}
