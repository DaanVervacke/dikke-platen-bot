package handlers

import (
	"dikkeplaten/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	SongLinkBaseURL     = "https://api.song.link/v1-alpha.1/links"
	SongLinkUserCountry = "BE"
)

var (
	re    = regexp.MustCompile(`https?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\\(),]|%[0-9a-fA-F][0-9a-fA-F])+`)
	hosts = map[string]struct{}{
		"open.spotify.com":   {},
		"www.spotify.com":    {},
		"music.youtube.com":  {},
		"www.youtube.com":    {},
		"youtu.be":           {},
		"soundcloud.com":     {},
		"www.soundcloud.com": {},
		"music.apple.com":    {},
		"www.apple.com":      {},
	}
)

func filterMessage(message string) (bool, string) {
	musicURLs := re.FindAllString(message, -1)

	if len(musicURLs) == 0 {
		return false, ""
	}

	u, err := url.Parse(musicURLs[0])
	if err != nil {
		return false, ""
	}

	host := strings.ToLower(u.Host)

	if _, ok := hosts[host]; ok {
		return true, u.String()
	}

	return false, ""
}

func buildSongLinkURL(musicURL string) (string, error) {
	baseURL, err := url.Parse(SongLinkBaseURL)
	if err != nil {
		return "", fmt.Errorf("error parsing base URL: %v", err)
	}

	params := url.Values{}
	params.Add("url", musicURL)
	params.Add("userCountry", SongLinkUserCountry)
	params.Add("songIfSingle", "true")

	baseURL.RawQuery = params.Encode()

	return baseURL.String(), nil
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

func GetSongLinkData(musicURL string) (types.FinalLinks, error) {
	songLinkURL, err := buildSongLinkURL(musicURL)
	if err != nil {
		return types.FinalLinks{}, err
	}

	resp, err := http.Get(songLinkURL)
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
