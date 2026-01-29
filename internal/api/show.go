package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Ka10ken1/mykadri-scraper/internal/models"
	"github.com/gin-gonic/gin"
)

func GetShows(c *gin.Context) {
	shows, err := models.GetAllShows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get shows"})
		return
	}
	c.JSON(http.StatusOK, shows)
}

func GetShowByID(c *gin.Context) {
	id := c.Param("id")

	show, err := models.GetShowByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get show"})
		return
	}

	if show == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "show not found"})
		return
	}

	c.JSON(http.StatusOK, show)
}

func GetShowImages(c *gin.Context) {
	images, err := models.GetAllShowImages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get show images"})
		return
	}
	c.JSON(http.StatusOK, images)
}

func ShowShowByTmdbID(c *gin.Context) {
	tmdbID := c.Param("tmdb_id")

	// 1. Check DB
	show, err := models.GetShowByTmdbID(tmdbID)
	if err == nil && show != nil {
		renderShow(c, show)
		return
	}

	// 2. Fetch from TMDB if missing
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		c.String(http.StatusInternalServerError, "TMDB_API_KEY not configured")
		return
	}

	url := fmt.Sprintf("https://api.themoviedb.org/3/tv/%s?api_key=%s", tmdbID, apiKey)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.String(http.StatusNotFound, "Show not found on TMDB")
		return
	}
	defer resp.Body.Close()

	var tmdbShow struct {
		Name       string `json:"name"`
		PosterPath string `json:"poster_path"`
		FirstAir   string `json:"first_air_date"`
	}

	// For TV, we need extra call for IMDB ID
	urlIDs := fmt.Sprintf("https://api.themoviedb.org/3/tv/%s/external_ids?api_key=%s", tmdbID, apiKey)
	respIDs, err := http.Get(urlIDs)
	imdbID := ""
	if err == nil && respIDs.StatusCode == http.StatusOK {
		var ids struct {
			ImdbID string `json:"imdb_id"`
		}
		json.NewDecoder(respIDs.Body).Decode(&ids)
		imdbID = ids.ImdbID
		respIDs.Body.Close()
	}

	json.NewDecoder(resp.Body).Decode(&tmdbShow)

	// 3. Create and Insert
	year := ""
	if len(tmdbShow.FirstAir) >= 4 {
		year = tmdbShow.FirstAir[:4]
	}

	newShow := models.Show{
		Title:        tmdbShow.Name,
		TitleEnglish: tmdbShow.Name,
		Year:         year,
		ImdbID:       imdbID,
		TmdbID:       tmdbID,
		Image:        fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", tmdbShow.PosterPath),
		VideoURL:     fmt.Sprintf("https://vidsrc-embed.ru/embed/tv/%s", imdbID),
	}

	if err := models.InsertShow(newShow); err != nil {
		log.Printf("Failed to cache show: %v", err)
	}

	renderShow(c, &newShow)
}

func renderShow(c *gin.Context, show *models.Show) {
	c.HTML(http.StatusOK, "show.html", gin.H{
		"Title":        show.Title,
		"EnglishTitle": show.TitleEnglish,
		"VideoURL":     show.VideoURL,
		"Image":        show.Image,
		"Year":         show.Year,
	})
}

func ShowShowPage(c *gin.Context) {
	id := c.Param("id")
	show, err := models.GetShowByID(id)
	if err != nil {
		c.String(http.StatusNotFound, "Show not found")
		return
	}

	renderShow(c, show)
}

func GetShowsByTitle(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	shows, err := models.SearchShowsByTitle(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search shows"})
		return
	}

	c.JSON(http.StatusOK, shows)
}
