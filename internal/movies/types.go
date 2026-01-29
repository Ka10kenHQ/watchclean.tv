package movies

type VidsrcMovieResponse struct {
	Result []VidsrcMovieItem `json:"result"`
	Pages  int               `json:"pages"`
}

type VidsrcMovieItem struct {
	ImdbID       string `json:"imdb_id"`
	TmdbID       string `json:"tmdb_id"`
	Title        string `json:"title"`
	EmbedURL     string `json:"embed_url"`
	EmbedURLTmdb string `json:"embed_url_tmdb"`
	Quality      string `json:"quality"`
}
