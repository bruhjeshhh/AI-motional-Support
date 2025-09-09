package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

const (
	kafkaBroker  = "localhost:9092"
	inputTopic   = "tweets"
	outputTopic  = "flagged-accounts"
	groupID      = "tweet-analyzer-group"
	offenseLimit = 3
)

var ctx = context.Background()

type Tweet struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

type FlaggedUser struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

func mockModel(text string) bool {
	suicideKeywords := []string{
		"feeling so alone",
		"don't know what to do with my life",
		"find a way out",
		"everything is so dark",
	}
	for _, keyword := range suicideKeywords {
		if strings.Contains(strings.ToLower(text), strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   inputTopic,
		GroupID: groupID,
	})
	defer reader.Close()

	writer := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    outputTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	var wg sync.WaitGroup
	log.Println("Tweet Analyzer service started. Waiting for messages...")

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		wg.Add(1)
		go func(m kafka.Message) {
			defer wg.Done()

			var tweet Tweet
			if err := json.Unmarshal(m.Value, &tweet); err != nil {
				log.Printf("Error unmarshaling tweet: %v", err)
				return
			}

			if mockModel(tweet.Text) {
				log.Printf("Detected suicidal content from user %s: '%s'", tweet.UserID, tweet.Text)

				offenses, err := rdb.Incr(ctx, tweet.UserID).Result()
				if err != nil {
					log.Printf("Redis error: %v", err)
					return
				}

				if offenses >= offenseLimit {
					alreadyFlagged, _ := rdb.Get(ctx, fmt.Sprintf("reported:%s", tweet.UserID)).Result()
					if alreadyFlagged == "" {
						flaggedUser := FlaggedUser{
							UserID: tweet.UserID,
							Reason: "Repetitive suicidal content",
							Count:  int(offenses),
						}
						flaggedJSON, _ := json.Marshal(flaggedUser)

						err = writer.WriteMessages(ctx, kafka.Message{
							Key:   []byte(tweet.UserID),
							Value: flaggedJSON,
						})
						if err != nil {
							log.Printf("Error writing flagged user: %v", err)
						}

						rdb.Set(ctx, fmt.Sprintf("reported:%s", tweet.UserID), "true", 24*time.Hour)
					} else {
						log.Printf("User %s already flagged. Skipping.", tweet.UserID)
					}
				}
			}
		}(m)
	}
	wg.Wait()
}
