package types

type LinksByPlatformItems struct {
	URL string `json:"url"`
}

type SongLinkResponse struct {
	LinksByPlatform map[string]LinksByPlatformItems `json:"linksByPlatform"`
}

type FinalLinks struct {
	Spotify      LinksByPlatformItems `json:"spotify"`
	Youtube      LinksByPlatformItems `json:"youtube"`
	YoutubeMusic LinksByPlatformItems `json:"youtubeMusic"`
	AppleMusic   LinksByPlatformItems `json:"appleMusic"`
	SoundCloud   LinksByPlatformItems `json:"soundcloud"`
}
