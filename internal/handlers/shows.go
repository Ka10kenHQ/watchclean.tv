package handlers

import (
	"log"

	"github.com/Ka10kenHQ/watchclean.tv/internal/models"
	"github.com/Ka10kenHQ/watchclean.tv/internal/shows"
)


func FetchShows() {
	hasShows, err := models.HasShows()
	if err != nil {
		log.Printf("Warning: Failed to check if shows exist: %v", err)
	}
	
	if hasShows {
		log.Println("Shows already exist in database, skipping fetch.")
		return
	}

	log.Println("Fetching show catalog (metadata from Vidsrc list, player URLs: Vidking)...")

	if err := shows.Fetch(); err != nil {
		log.Fatal(err)
	}

	log.Println("All shows processed and inserted.")
}
