package vidking

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	// EmbedBase is the Vidking player origin used for iframe src URLs.
	EmbedBase = "https://www.vidking.net"

	// DefaultTVSeason and DefaultTVEpisode are used when a catalog item has no
	// specific episode (Vidking requires season and episode in the TV path).
	DefaultTVSeason  = 1
	DefaultTVEpisode = 1
)

// EmbedOptions configures optional Vidking URL query parameters.
// Omitted bool parameters default to off in the player; only true values are appended.
type EmbedOptions struct {
	Color            string
	AutoPlay         bool
	NextEpisode      bool
	EpisodeSelector  bool
	ProgressSeconds  int // playback start offset; <= 0 is omitted
}

// MovieEmbedURL returns a Vidking movie iframe source URL using TMDB movie ID.
// Empty tmdbID yields an empty string.
func MovieEmbedURL(tmdbID string, o EmbedOptions) string {
	if tmdbID == "" {
		return ""
	}
	base := fmt.Sprintf("%s/embed/movie/%s", EmbedBase, url.PathEscape(tmdbID))
	return withQuery(base, o)
}

// TVEmbedURL returns a Vidking TV iframe source URL using TMDB show ID, season, and episode.
// Empty tmdbID yields an empty string.
func TVEmbedURL(tmdbID string, season, episode int, o EmbedOptions) string {
	if tmdbID == "" {
		return ""
	}
	base := fmt.Sprintf("%s/embed/tv/%s/%d/%d", EmbedBase, url.PathEscape(tmdbID), season, episode)
	return withQuery(base, o)
}

func withQuery(raw string, o EmbedOptions) string {
	if o.Color == "" && !o.AutoPlay && !o.NextEpisode && !o.EpisodeSelector && o.ProgressSeconds <= 0 {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	if o.Color != "" {
		q.Set("color", o.Color)
	}
	if o.AutoPlay {
		q.Set("autoPlay", "true")
	}
	if o.NextEpisode {
		q.Set("nextEpisode", "true")
	}
	if o.EpisodeSelector {
		q.Set("episodeSelector", "true")
	}
	if o.ProgressSeconds > 0 {
		q.Set("progress", strconv.Itoa(o.ProgressSeconds))
	}
	u.RawQuery = q.Encode()
	return u.String()
}
