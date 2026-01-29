package movies

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Ka10ken1/mykadri-scraper/internal/common"
	"github.com/Ka10ken1/mykadri-scraper/internal/models"
)

type Movie = models.Movie

func Fetch() error {
	alreadyFetched, err := models.HasMovies()
	if err != nil {
		return fmt.Errorf("db check error: %w", err)
	}
	if alreadyFetched {
		log.Println("Movies already fetched, skipping.")
		return nil
	}

	existingImdbIds, err := models.GetAllMovieImdbIds()
	if err != nil {
		return fmt.Errorf("failed to preload movie IDs: %w", err)
	}

	seen := make(map[string]struct{}, len(existingImdbIds))
	for _, id := range existingImdbIds {
		seen[id] = struct{}{}
	}

	firstPage, err := fetchPage(1)
	if err != nil {
		return fmt.Errorf("failed to fetch first page: %w", err)
	}

	totalPages := firstPage.Pages
	log.Printf("Total movie pages available: %d", totalPages)

	maxPages := totalPages

	var wg sync.WaitGroup
	sema := make(chan struct{}, common.MovieParallelism)

	var mu sync.Mutex

	processPage := func(pageData *VidsrcMovieResponse) {
		var batch []models.Movie

		for _, item := range pageData.Result {
			mu.Lock()
			_, exists := seen[item.ImdbID]
			if !exists {
				seen[item.ImdbID] = struct{}{}
			}
			mu.Unlock()

			if !exists {
				movie := parseFromAPI(item)
				batch = append(batch, movie)
			}
		}

		if len(batch) > 0 {
			if err := models.InsertMovies(batch); err != nil {
				log.Printf("Failed to insert batch: %v", err)
			} else {
				log.Printf("Inserted batch of %d movies", len(batch))
			}
		}
	}

	// Insert first page immediately
	processPage(firstPage)

	for i := 2; i <= maxPages; i++ {
		wg.Add(1)
		sema <- struct{}{}

		go func(page int) {
			defer func() {
				<-sema
				wg.Done()
			}()

			pageData, err := fetchPage(page)
			if err != nil {
				log.Printf("Failed to fetch page %d: %v", page, err)
				return
			}

			processPage(pageData)
		}(i)
	}

	wg.Wait()
	log.Println("All pages processed.")
	return nil
}

func fetchPage(page int) (*VidsrcMovieResponse, error) {
	url := fmt.Sprintf("https://vidsrc-embed.ru/movies/latest/page-%d.json", page)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var movieResp VidsrcMovieResponse
	if err := json.Unmarshal(body, &movieResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &movieResp, nil
}

func parseFromAPI(item VidsrcMovieItem) Movie {
	year := common.ExtractYear(item.Title)
	cleanTitle := common.RemoveYearFromTitle(item.Title)
	embedURL := common.ConvertToNewEmbedURL(item.EmbedURL)

	imgURL := ""
	tmdbKey := os.Getenv("TMDB_API_KEY")
	if item.TmdbID != "" && tmdbKey != "" {
		imgURL = getPosterFromTMDB(item.TmdbID, "movie", tmdbKey)
	}

	if imgURL == "" && item.TmdbID != "" {
		imgURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500/%s.jpg", item.TmdbID)
	}

	return Movie{
		Title:        cleanTitle,
		TitleEnglish: cleanTitle,
		Year:         year,
		ImdbID:       item.ImdbID,
		TmdbID:       item.TmdbID,
		Image:        imgURL,
		VideoURL:     embedURL,
		Quality:      item.Quality,
	}
}

func getPosterFromTMDB(tmdbID, mediaType, apiKey string) string {
	url := fmt.Sprintf("https://api.themoviedb.org/3/%s/%s?api_key=%s", mediaType, tmdbID, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result struct {
		PosterPath string `json:"poster_path"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	if result.PosterPath != "" {
		return fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", result.PosterPath)
	}
	return ""
}
