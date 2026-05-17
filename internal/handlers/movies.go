package handlers

import (
	"log"

	"github.com/Ka10kenHQ/watchclean.tv/internal/models"
	"github.com/Ka10kenHQ/watchclean.tv/internal/movies"
)


func FetchMovies() {
	hasMovies, err := models.HasMovies()
	if err != nil {
		log.Printf("Warning: Failed to check if movies exist: %v", err)
	}
	
	if hasMovies {
		log.Println("Movies already exist in database, skipping fetch.")
		return
	}

	log.Println("Fetching movie catalog (metadata from Vidsrc list, player URLs: Vidking)...")

	if err := movies.Fetch(); err != nil {
		log.Fatal(err)
	}

	log.Println("All movies processed and inserted.")
}

