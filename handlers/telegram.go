package handlers

import (
	"dikkeplaten/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func buildWebhookURL(baseURL *url.URL, webhookURL string, webhookSecret string) (string, error) {
	baseURL = baseURL.JoinPath("setWebhook")

	allowedUpdates, _ := json.Marshal([]string{"channel_post"})

	params := url.Values{}
	params.Add("url", webhookURL)
	params.Add("allowed_updates", string(allowedUpdates))
	params.Add("secret_token", webhookSecret)

	baseURL.RawQuery = params.Encode()

	return baseURL.String(), nil
}

func buildDeleteMessageURL(baseURL *url.URL, update types.TelegramBotUpdate) (string, error) {
	baseURL = baseURL.JoinPath("deleteMessage")

	params := url.Values{}
	params.Add("chat_id", strconv.Itoa(update.Message.Chat.ID))
	params.Add("message_id", strconv.Itoa(update.Message.ID))

	baseURL.RawQuery = params.Encode()

	return baseURL.String(), nil
}

func buildEditMessageURL(baseURL *url.URL, update types.TelegramBotUpdate, songLinkData types.FinalLinks) (string, error) {
	baseURL = baseURL.JoinPath("sendMessage")

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

func SetWebhook(baseURL *url.URL, webhookUrl string, webhookSecret string) error {
	webhookURL, err := buildWebhookURL(baseURL, webhookUrl, webhookSecret)
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

func updateMessage(baseURL *url.URL, update types.TelegramBotUpdate, songLinkData types.FinalLinks) error {
	deleteURL, err := buildDeleteMessageURL(baseURL, update)
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

	editUrl, err := buildEditMessageURL(baseURL, update, songLinkData)
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

func HandleBotUpdate(r *http.Request, baseURL *url.URL, groupID int, webhookSecret string, songLinkBaseURL *url.URL) error {
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

	ok, songURL := filterMessage(update.Message.Text)
	if !ok {
		return fmt.Errorf("invalid song url")
	}

	songLinkdata, err := GetSongLinkData(songLinkBaseURL, songURL)
	if err != nil {
		return err
	}

	err = updateMessage(baseURL, update, songLinkdata)
	if err != nil {
		return err
	}

	return nil
}
