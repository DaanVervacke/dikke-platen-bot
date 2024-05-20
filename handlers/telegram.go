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

func buildTelegramWebhookUrl(telegramBaseUrl string, webhookUrl string) (string, error) {
	baseUrl, err := url.Parse(telegramBaseUrl)
	if err != nil {
		return "", fmt.Errorf("error parsing base URL: %v", err)
	}

	baseUrl.Path = path.Join(baseUrl.Path, "setWebhook")

	params := url.Values{}
	params.Add("url", webhookUrl)

	allowedUpdates, _ := json.Marshal([]string{"message"})
	params.Add("allowed_updates", string(allowedUpdates))

	baseUrl.RawQuery = params.Encode()

	return baseUrl.String(), nil
}

func buildTelegramDeleteMessageUrl(telegramBaseUrl string, update types.TelegramBotUpdate) (string, error) {
	baseUrl, err := url.Parse(telegramBaseUrl)
	if err != nil {
		return "", fmt.Errorf("error parsing base URL: %v", err)
	}

	baseUrl.Path = path.Join(baseUrl.Path, "deleteMessage")

	params := url.Values{}
	params.Add("chat_id", strconv.Itoa(update.Message.Chat.ID))
	params.Add("message_id", strconv.Itoa(update.Message.ID))

	baseUrl.RawQuery = params.Encode()

	return baseUrl.String(), nil
}

func buildTelegramEditMessageUrl(telegramBaseUrl string, update types.TelegramBotUpdate, songLinkData types.FinalLinks) (string, error) {
	baseUrl, err := url.Parse(telegramBaseUrl)
	if err != nil {
		return "", fmt.Errorf("error parsing base URL: %v", err)
	}

	baseUrl.Path = path.Join(baseUrl.Path, "sendMessage")

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

	baseUrl.RawQuery = params.Encode()

	return baseUrl.String(), nil
}

func SetTelegramWebhook(telegramBaseUrl string, webhookUrl string) error {
	url, err := buildTelegramWebhookUrl(telegramBaseUrl, webhookUrl)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func updateTelegramMessage(telegramBaseUrl string, update types.TelegramBotUpdate, songLinkData types.FinalLinks) error {
	deleteUrl, err := buildTelegramDeleteMessageUrl(telegramBaseUrl, update)
	if err != nil {
		return err
	}

	resp, err := http.Post(deleteUrl, "application/json", nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("unexpected status code when deleting message: %d", resp.StatusCode)
	}

	resp.Body.Close()

	editUrl, err := buildTelegramEditMessageUrl(telegramBaseUrl, update, songLinkData)
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

func HandleBotUpdate(r *http.Request, telegramBaseUrl string, groupId int) error {

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

	if update.Message.Chat.ID != groupId {
		return fmt.Errorf("invalid group id: %d", update.Message.Chat.ID)
	}

	if !filterMessages(update.Message.Text) {
		return fmt.Errorf("invalid url: %s", update.Message.Text)
	}

	songLinkdata, err := GetSongLinkData(update.Message.Text)
	if err != nil {
		return err
	}

	err = updateTelegramMessage(telegramBaseUrl, update, songLinkdata)
	if err != nil {
		return err
	}

	return nil

}
