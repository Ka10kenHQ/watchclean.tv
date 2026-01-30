package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Ka10kenHQ/watchclean.tv/internal/models"
	"github.com/gin-gonic/gin"
)

func GetMovies(c *gin.Context) {
	movies, err := models.GetAllMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get movies"})
		return
	}
	c.JSON(http.StatusOK, movies)
}

func GetMovieByID(c *gin.Context) {
	id := c.Param("id")

	movie, err := models.GetMovieByID(id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get movie"})
		return
	}

	if movie == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}

	c.JSON(http.StatusOK, movie)
}

func GetMovieImages(c *gin.Context) {
	images, err := models.GetAllMovieImages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get movie images"})
		return
	}
	c.JSON(http.StatusOK, images)
}

func ShowMoviePage(c *gin.Context) {
	id := c.Param("id")
	movie, err := models.GetMovieByID(id)
	if err != nil {
		c.String(http.StatusNotFound, "Movie not found")
		return
	}

	renderMovie(c, movie)
}

func ShowMovieByTmdbID(c *gin.Context) {
	tmdbID := c.Param("tmdb_id")

	// 1. Check DB
	movie, err := models.GetMovieByTmdbID(tmdbID)
	if err == nil && movie != nil {
		renderMovie(c, movie)
		return
	}

	// 2. Fetch from TMDB if missing
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		c.String(http.StatusInternalServerError, "TMDB_API_KEY not configured")
		return
	}

	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%s?api_key=%s", tmdbID, apiKey)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.String(http.StatusNotFound, "Movie not found on TMDB")
		return
	}
	defer resp.Body.Close()

	var tmdbMovie struct {
		Title      string `json:"title"`
		PosterPath string `json:"poster_path"`
		Release    string `json:"release_date"`
		ImdbID     string `json:"imdb_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tmdbMovie); err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse TMDB data")
		return
	}

	// 3. Create and Insert
	year := ""
	if len(tmdbMovie.Release) >= 4 {
		year = tmdbMovie.Release[:4]
	}

	newMovie := models.Movie{
		Title:        tmdbMovie.Title,
		TitleEnglish: tmdbMovie.Title,
		Year:         year,
		ImdbID:       tmdbMovie.ImdbID,
		TmdbID:       tmdbID,
		Image:        fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", tmdbMovie.PosterPath),
		VideoURL:     fmt.Sprintf("https://vidsrc-embed.ru/embed/movie/%s", tmdbMovie.ImdbID),
	}

	if err := models.InsertMovie(newMovie); err != nil {
		log.Printf("Failed to cache movie: %v", err)
	}

	renderMovie(c, &newMovie)
}

func renderMovie(c *gin.Context, movie *models.Movie) {
	c.HTML(http.StatusOK, "movie.html", gin.H{
		"Title":        movie.Title,
		"EnglishTitle": movie.TitleEnglish,
		"VideoURL":     movie.VideoURL,
		"Image":        movie.Image,
		"Year":         movie.Year,
	})
}

func GetMoviesByTitle(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	movies, err := models.SearchMoviesByTitle(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search movies"})
		return
	}

	c.JSON(http.StatusOK, movies)
}
