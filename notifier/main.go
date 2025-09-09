package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
)

const (
	kafkaBroker = "localhost:9092"
	inputTopic  = "flagged-accounts"
	groupID     = "twilio-notifier-group"
)

var ctx = context.Background()

type FlaggedUser struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

func sendSMS(client *twilio.RestClient, to, body string) {
	fromNumber := os.Getenv("TWILIO_FROM_NUMBER")
	if fromNumber == "" {
		log.Println("Twilio FROM number not set. Skipping SMS.")
		return
	}

	params := &api.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(fromNumber)
	params.SetBody(body)

	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		log.Printf("Error sending SMS: %v", err)
		return
	}

	log.Printf("SMS sent. SID: %s", *resp.Sid)
}

func main() {
	client := twilio.NewRestClient()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   inputTopic,
		GroupID: groupID,
	})
	defer reader.Close()

	log.Println("Twilio Notifier started. Listening for flagged users...")

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var flagged FlaggedUser
		if err := json.Unmarshal(m.Value, &flagged); err != nil {
			log.Printf("Error unmarshaling flagged user: %v", err)
			continue
		}

		toPhone := os.Getenv("PERSONAL_PHONE_NUMBER")
		if toPhone == "" {
			log.Println("PERSONAL_PHONE_NUMBER not set. Skipping SMS.")
			continue
		}

		msg := fmt.Sprintf("ALERT 🚨 User %s flagged for %d suicidal posts. Reason: %s",
			flagged.UserID, flagged.Count, flagged.Reason)

		sendSMS(client, toPhone, msg)
	}
}
