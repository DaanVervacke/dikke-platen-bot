package handlers

import (
	"dikkeplaten/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

func buildTelegramWebhookUrl(telegramBaseURL string, webhookURL string, webhookSecret string) (string, error) {
	baseURL, err := url.Parse(telegramBaseURL)
	if err != nil {
		return "", fmt.Errorf("error parsing base URL: %v", err)
	}

	baseURL.Path = path.Join(baseURL.Path, "setWebhook")

	params := url.Values{}
	params.Add("url", webhookURL)

	allowedUpdates, _ := json.Marshal([]string{"channel_post"})
	params.Add("allowed_updates", string(allowedUpdates))

	params.Add("secret_token", webhookSecret)

	baseURL.RawQuery = params.Encode()

	return baseURL.String(), nil
}

func buildTelegramDeleteMessageUrl(telegramBaseURL string, update types.TelegramBotUpdate) (string, error) {
	baseURL, err := url.Parse(telegramBaseURL)
	if err != nil {
		return "", fmt.Errorf("error parsing base URL: %v", err)
	}

	baseURL.Path = path.Join(baseURL.Path, "deleteMessage")

	params := url.Values{}
	params.Add("chat_id", strconv.Itoa(update.Message.Chat.ID))
	params.Add("message_id", strconv.Itoa(update.Message.ID))

	baseURL.RawQuery = params.Encode()

	return baseURL.String(), nil
}

func buildTelegramEditMessageUrl(telegramBaseURL string, update types.TelegramBotUpdate, songLinkData types.FinalLinks) (string, error) {
	baseURL, err := url.Parse(telegramBaseURL)
	if err != nil {
		return "", fmt.Errorf("error parsing base URL: %v", err)
	}

	baseURL.Path = path.Join(baseURL.Path, "sendMessage")

	params := url.Values{}
	params.Add("chat_id", strconv.Itoa(update.Message.Chat.ID))

	var text strings.Builder

	if songLinkData.Spotify.URL != "" {
		text.WriteString(fmt.Sprintf("Spotify: %s\n\n", songLinkData.Spotify.URL))
	}

	if songLinkData.Youtube.URL != "" {
		text.WriteString(fmt.Sprintf("Youtube: %s\n\n", songLinkData.Youtube.URL))
	}

	if songLinkData.YoutubeMusic.URL != "" {
		text.WriteString(fmt.Sprintf("Youtube Music: %s\n\n", songLinkData.YoutubeMusic.URL))
	}

	if songLinkData.AppleMusic.URL != "" {
		text.WriteString(fmt.Sprintf("Apple Music: %s\n\n", songLinkData.AppleMusic.URL))
	}

	if songLinkData.SoundCloud.URL != "" {
		text.WriteString(fmt.Sprintf("SoundCloud: %s\n\n", songLinkData.SoundCloud.URL))
	}

	params.Add("text", text.String())

	baseURL.RawQuery = params.Encode()

	return baseURL.String(), nil
}

func SetTelegramWebhook(telegramBaseUrl string, webhookUrl string, webhookSecret string) error {
	webhookURL, err := buildTelegramWebhookUrl(telegramBaseUrl, webhookUrl, webhookSecret)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func updateTelegramMessage(telegramBaseURL string, update types.TelegramBotUpdate, songLinkData types.FinalLinks) error {
	deleteURL, err := buildTelegramDeleteMessageUrl(telegramBaseURL, update)
	if err != nil {
		return err
	}

	resp, err := http.Post(deleteURL, "application/json", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("unexpected status code when deleting message: %d", resp.StatusCode)
	}

	resp.Body.Close()

	editUrl, err := buildTelegramEditMessageUrl(telegramBaseURL, update, songLinkData)
	if err != nil {
		return err
	}

	resp, err = http.Post(editUrl, "application/json", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("unexpected status code when editing message: %d", resp.StatusCode)
	}

	resp.Body.Close()

	return nil
}

func HandleBotUpdate(r *http.Request, telegramBaseURL string, groupID int, webhookSecret string) error {
	if r.Header.Get("X-Telegram-Bot-Api-Secret-Token") != webhookSecret {
		return fmt.Errorf("invalid secret token")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	var update types.TelegramBotUpdate

	err = json.Unmarshal(body, &update)
	if err != nil {
		return err
	}

	if update.Message.Chat.ID != groupID {
		return fmt.Errorf("invalid group id: %d", update.Message.Chat.ID)
	}

	ok, musicURL := filterMessage(update.Message.Text)
	if !ok {
		return fmt.Errorf("invalid url")
	}

	songLinkdata, err := GetSongLinkData(musicURL)
	if err != nil {
		return err
	}

	err = updateTelegramMessage(telegramBaseURL, update, songLinkdata)
	if err != nil {
		return err
	}

	return nil
}
