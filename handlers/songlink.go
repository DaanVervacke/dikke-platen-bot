package handlers

import (
	"dikkeplaten/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	SongLinkBaseURL     = "https://api.song.link/v1-alpha.1/links"
	SongLinkUserCountry = "BE"
)

func filterMessages(message string) bool {
	u, err := url.Parse(message)
	if err != nil {
		return false
	}

	host := strings.ToLower(u.Host)

	switch host {
	case "open.spotify.com", "www.spotify.com", "music.youtube.com", "www.youtube.com", "youtu.be", "soundcloud.com", "www.soundcloud.com", "music.apple.com", "www.apple.com":
		return true
	default:
		return false
	}
}

func buildSongLinkUrl(musicUrl string) (string, error) {
	baseUrl, err := url.Parse(SongLinkBaseURL)
	if err != nil {
		return "", fmt.Errorf("error parsing base URL: %v", err)
	}

	params := url.Values{}
	params.Add("url", musicUrl)
	params.Add("userCountry", SongLinkUserCountry)
	params.Add("songIfSingle", "true")

	baseUrl.RawQuery = params.Encode()

	return baseUrl.String(), nil
}

func parseResponse(response *http.Response) (types.FinalLinks, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return types.FinalLinks{}, err
	}

	var songLinkResponse types.SongLinkResponse

	err = json.Unmarshal(body, &songLinkResponse)
	if err != nil {
		return types.FinalLinks{}, err
	}

	return types.FinalLinks{
		Spotify:      songLinkResponse.LinksByPlatform["spotify"],
		Youtube:      songLinkResponse.LinksByPlatform["youtube"],
		YoutubeMusic: songLinkResponse.LinksByPlatform["youtubeMusic"],
		AppleMusic:   songLinkResponse.LinksByPlatform["appleMusic"],
		SoundCloud:   songLinkResponse.LinksByPlatform["soundcloud"],
	}, nil
}

func GetSongLinkData(musicUrl string) (types.FinalLinks, error) {
	url, err := buildSongLinkUrl(musicUrl)
	if err != nil {
		return types.FinalLinks{}, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return types.FinalLinks{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.FinalLinks{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	finalLinks, err := parseResponse(resp)
	if err != nil {
		return types.FinalLinks{}, err
	}

	return finalLinks, nil
}