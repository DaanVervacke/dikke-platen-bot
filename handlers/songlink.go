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

var (
	re    = regexp.MustCompile(`https?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\\(),]|%[0-9a-fA-F][0-9a-fA-F])+`)
	hosts = map[string]struct{}{
		"open.spotify.com":   {},
		"www.spotify.com":    {},
		"music.youtube.com":  {},
		"www.youtube.com":    {},
		"youtu.be":           {},
		"soundcloud.com":     {},
		"on.soundcloud.com":  {},
		"www.soundcloud.com": {},
		"music.apple.com":    {},
		"www.apple.com":      {},
	}
)

func filterMessage(message string) (bool, string) {
	songURLs := re.FindAllString(message, -1)

	if len(songURLs) == 0 {
		return false, ""
	}

	u, err := url.Parse(songURLs[0])
	if err != nil {
		return false, ""
	}

	host := strings.ToLower(u.Host)

	if _, ok := hosts[host]; ok {
		return true, u.String()
	}

	return false, ""
}

func buildURL(baseURL *url.URL, musicURL string) (string, error) {
	params := url.Values{}
	params.Add("url", musicURL)
	params.Add("userCountry", "BE")
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

func GetSongLinkData(baseURL *url.URL, songURL string) (types.FinalLinks, error) {
	songLinkURL, err := buildURL(baseURL, songURL)
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

	links, err := parseResponse(resp)
	if err != nil {
		return types.FinalLinks{}, err
	}

	return links, nil
}
