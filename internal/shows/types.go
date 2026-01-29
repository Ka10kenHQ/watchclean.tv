package shows

type VidsrcShowResponse struct {
	Result []VidsrcShowItem `json:"result"`
	Pages  int              `json:"pages"`
}

type VidsrcShowItem struct {
	ImdbID       string  `json:"imdb_id"`
	TmdbID       *string `json:"tmdb_id"`
	Title        string  `json:"title"`
	EmbedURL     string  `json:"embed_url"`
	EmbedURLTmdb string  `json:"embed_url_tmdb,omitempty"`
}
