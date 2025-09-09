package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

const (
	kafkaBroker = "localhost:9092"
	topic       = "tweets"
)

type Tweet struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

func initLogger() {
	// log format: DATE TIME | message
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)
}

func generateFakeTweets() {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	offensiveTweets := []string{
		"I'm feeling so alone, nothing seems to matter anymore.",
		"Feeling hopeless, don't know what to do with my life.",
		"Everything is so dark, I can't find a way out.",
	}
	nonOffensiveTweets := []string{
		"What a beautiful day! The sun is shining.",
		"Just finished my project, feeling great!",
		"Enjoying a cup of coffee. Simple pleasures.",
		"Had a fun time with friends today.",
	}
	offendingUsers := []string{"offender-1", "offender-2", "offender-3"}

	log.Println("[Producer] Starting continuous tweet generation...")

	for i := 0; ; i++ {
		var userID, text string
		var tweetType string

		if i%20 == 0 {
			offenderIndex := (i / 20) % len(offendingUsers)
			userID = offendingUsers[offenderIndex]
			text = offensiveTweets[offenderIndex%len(offensiveTweets)]
			tweetType = "OFFENSIVE"
		} else {
			userID = fmt.Sprintf("user-%d", i%10000)
			text = nonOffensiveTweets[i%len(nonOffensiveTweets)]
			tweetType = "NORMAL"
		}

		tweet := Tweet{
			ID:        fmt.Sprintf("tweet-%d", i),
			UserID:    userID,
			Text:      text,
			Timestamp: time.Now(),
		}

		tweetJSON, err := json.Marshal(tweet)
		if err != nil {
			log.Printf("[Producer] ❌ Error marshaling tweet: %v", err)
			continue
		}

		err = writer.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(userID),
				Value: tweetJSON,
			},
		)
		if err != nil {
			log.Printf("[Producer] ❌ Error writing message to Kafka: %v", err)
		} else {
			log.Printf("[Producer] ✅ Sent %s tweet | ID: %s | User: %s | Text: %s",
				tweetType, tweet.ID, tweet.UserID, tweet.Text)
		}

		if (i+1)%1000 == 0 {
			log.Printf("[Producer] 🚀 Produced %d tweets so far...", i+1)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	initLogger()
	generateFakeTweets()
}
