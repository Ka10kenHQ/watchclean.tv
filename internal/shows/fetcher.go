package shows

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

type Show = models.Show

func Fetch() error {
	alreadyFetched, err := models.HasShows()
	if err != nil {
		return fmt.Errorf("db check error: %w", err)
	}
	if alreadyFetched {
		log.Println("Shows already fetched, skipping.")
		return nil
	}

	existingImdbIds, err := models.GetAllShowImdbIds()
	if err != nil {
		return fmt.Errorf("failed to preload show IDs: %w", err)
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
	log.Printf("Total show pages available: %d", totalPages)

	maxPages := totalPages

	var wg sync.WaitGroup
	sema := make(chan struct{}, common.ShowParallelism)
	var mu sync.Mutex

	processPage := func(pageData *VidsrcShowResponse) {
		var batch []models.Show

		for _, item := range pageData.Result {

			mu.Lock()
			_, exists := seen[item.ImdbID]
			if !exists {
				seen[item.ImdbID] = struct{}{}
			}
			mu.Unlock()

			if !exists {
				show := parseFromAPI(item)
				batch = append(batch, show)
			}
		}

		if len(batch) > 0 {
			if err := models.InsertShows(batch); err != nil {
				log.Printf("Failed to insert batch: %v", err)
			} else {
				log.Printf("Inserted batch of %d shows", len(batch))
			}
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		processPage(firstPage)
	}()

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
	log.Println("All show pages processed.")
	return nil
}

func fetchPage(page int) (*VidsrcShowResponse, error) {
	url := fmt.Sprintf("https://vidsrc-embed.ru/tvshows/latest/page-%d.json", page)

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

	var showResp VidsrcShowResponse
	if err := json.Unmarshal(body, &showResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &showResp, nil
}

func parseFromAPI(item VidsrcShowItem) Show {
	year := common.ExtractYear(item.Title)
	cleanTitle := common.RemoveYearFromTitle(item.Title)
	embedURL := common.ConvertToNewEmbedURL(item.EmbedURL)

	tmdbID := ""
	if item.TmdbID != nil {
		tmdbID = *item.TmdbID
	}

	var imgURL string
	tmdbKey := os.Getenv("TMDB_API_KEY")
	if tmdbID != "" && tmdbKey != "" {
		imgURL = getPosterFromTMDB(tmdbID, "tv", tmdbKey)
	}

	if imgURL == "" && tmdbID != "" {
		imgURL = fmt.Sprintf("https://image.tmdb.org/t/p/w500/%s.jpg", tmdbID)
	}

	return Show{
		Title:        cleanTitle,
		TitleEnglish: cleanTitle,
		Year:         year,
		ImdbID:       item.ImdbID,
		TmdbID:       tmdbID,
		Image:        imgURL,
		VideoURL:     embedURL,
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
