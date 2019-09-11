// Package main contains the main logic to receive triggers from AWS CloudWatch
// and get Trello cards based on that to send to Slack
package main

import (
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/kelseyhightower/envconfig"
	"github.com/retgits/weekly-review/slack"
	"github.com/retgits/weekly-review/trello"
)

type config struct {
	AWSRegion        string `required:"true" split_words:"true" envconfig:"AWS_REGION"`
	TrelloAppKey     string `required:"true" split_words:"true"`
	TrelloAppToken   string `required:"true" split_words:"true"`
	TrelloMemberName string `required:"true" split_words:"true"`
	TrelloBoard      string `required:"true" split_words:"true"`
	TrelloList       string `required:"true" split_words:"true"`
	SlackWebhookURL  string `required:"true" split_words:"true"`
	SlackChannel     string `required:"true" split_words:"true"`
	SlackUser        string `required:"true" split_words:"true"`
	SlackEmoji       string `required:"true" split_words:"true"`
}

var c config

func handler(request events.CloudWatchEvent) error {
	// Get configuration set using environment variables
	err := envconfig.Process("", &c)
	if err != nil {
		return err
	}

	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(c.AWSRegion),
	}))

	kmsSvc := kms.New(awsSession)

	val, err := decodeString(kmsSvc, c.TrelloAppKey)
	if err != nil {
		return err
	}
	c.TrelloAppKey = val

	val, err = decodeString(kmsSvc, c.TrelloAppToken)
	if err != nil {
		return err
	}
	c.TrelloAppToken = val

	val, err = decodeString(kmsSvc, c.TrelloMemberName)
	if err != nil {
		return err
	}
	c.TrelloMemberName = val

	val, err = decodeString(kmsSvc, c.SlackWebhookURL)
	if err != nil {
		return err
	}
	c.SlackWebhookURL = val

	trelloSvc := trello.New().WithKey(c.TrelloAppKey).WithToken(c.TrelloAppToken).WithMemberName(c.TrelloMemberName)

	gci := &trello.GetCardInput{
		Board:      c.TrelloBoard,
		List:       c.TrelloList,
		WithLabels: true,
	}

	gco, err := trelloSvc.GetCards(gci)
	if err != nil {
		return err
	}

	slackSvc := slack.New().WithWebhook(c.SlackWebhookURL)

	slackAttachments := make([]slack.Attachment, len(gco.Cards)+1)

	slackAttachments[0] = slack.Attachment{
		Title: "Your weekly review",
		Text:  fmt.Sprintf("These are the cards that appear in your %s list this week", c.TrelloList),
		Color: "0C31EA",
	}

	for idx, card := range gco.Cards {
		slackAttachments[idx+1] = slack.Attachment{
			Fallback: card.Name,
			Title:    fmt.Sprintf("%s: %+v", card.Name, card.Labels),
			Text:     card.Description,
			Color:    "4B0040",
		}
	}

	slackMsg := &slack.Message{
		Channel:     c.SlackChannel,
		Username:    c.SlackUser,
		IconEmoji:   fmt.Sprintf(":%s:", c.SlackEmoji),
		Attachments: slackAttachments,
	}

	return slackSvc.Send(slackMsg)
}

// decodeString uses AWS Key Management Service (AWS KMS) to decrypt environment variables.
// In order for this method to work, the function needs access to the kms:Decrypt capability.
func decodeString(kmsSvc *kms.KMS, payload string) (string, error) {
	sDec, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", err
	}
	out, err := kmsSvc.Decrypt(&kms.DecryptInput{
		CiphertextBlob: sDec,
	})
	if err != nil {
		return "", err
	}
	return string(out.Plaintext), nil
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(handler)
}
