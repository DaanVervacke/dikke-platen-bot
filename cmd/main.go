package main

import (
	"dikkeplaten/handlers"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	TelegramBaseURL = "https://api.telegram.org/bot"
)

func getEnvVar(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Missing required environment variable %s", key)
	}
	return value
}

func main() {
	telegramAPIToken := getEnvVar("TELEGRAM_API_TOKEN")
	webhookURL := getEnvVar("WEBHOOK_URL")
	webhookSecret := getEnvVar("WEBHOOK_SECRET")

	telegramGroupIDStr := getEnvVar("TELEGRAM_GROUP_ID")

	telegramGroupID, err := strconv.Atoi(telegramGroupIDStr)
	if err != nil {
		log.Fatalf("Error converting TELEGRAM_GROUP_ID to int: %v", err)
	}

	baseURL := fmt.Sprintf("%s%s", TelegramBaseURL, telegramAPIToken)

	router := http.NewServeMux()
	router.HandleFunc("POST /dikkeplaten", func(w http.ResponseWriter, r *http.Request) {
		err := handlers.HandleBotUpdate(r, baseURL, telegramGroupID, webhookSecret)
		if err != nil {
			log.Print(err)
		}
	})
	router.HandleFunc("GET /healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	err = handlers.SetTelegramWebhook(baseURL, webhookURL, webhookSecret)
	if err != nil {
		log.Fatalf("Error setting Telegram webhook: %v", err)
	}

	log.Fatal(http.ListenAndServe(":8080", router))
}
