package main

import (
	"dikkeplaten/handlers"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func getEnvVar(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Missing required environment variable %s", key)
	}
	return value
}

func main() {
	webhookURL := getEnvVar("WEBHOOK_URL")
	webhookSecret := getEnvVar("WEBHOOK_SECRET")

	groupID, err := strconv.Atoi(getEnvVar("TELEGRAM_GROUP_ID"))
	if err != nil {
		log.Fatalf("Error converting TELEGRAM_GROUP_ID to int: %v", err)
	}

	telegramBaseURL, err := url.Parse(fmt.Sprintf("%s%s", getEnvVar("TELEGRAM_BOT_URL"), getEnvVar("TELEGRAM_API_TOKEN")))
	if err != nil {
		log.Fatalf("Error parsing Telegram Bot URL: %v", err)
	}

	songlinkBaseURL, err := url.Parse(getEnvVar("SONGLINK_BASE_URL"))
	if err != nil {
		log.Fatalf("Error parsing Songlink URL: %v", err)
	}

	router := http.NewServeMux()

	router.HandleFunc("POST /dikkeplaten", func(w http.ResponseWriter, r *http.Request) {
		err := handlers.HandleBotUpdate(r, telegramBaseURL, groupID, webhookSecret, songlinkBaseURL)
		if err != nil {
			log.Print(err)
		}
	})

	router.HandleFunc("GET /healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	err = handlers.SetWebhook(telegramBaseURL, webhookURL, webhookSecret)
	if err != nil {
		log.Fatalf("Error setting Telegram webhook: %v", err)
	}

	log.Fatal(http.ListenAndServe(":8888", router))
}
