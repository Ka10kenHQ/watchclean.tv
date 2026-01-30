package main

import (
	"log"
	"os"

	"github.com/Ka10kenHQ/watchclean.tv/internal/api"
	"github.com/Ka10kenHQ/watchclean.tv/internal/models"
	"github.com/Ka10kenHQ/watchclean.tv/internal/movies"
	"github.com/Ka10kenHQ/watchclean.tv/internal/shows"
	"github.com/joho/godotenv"
)


func fetchMovies() {
	hasMovies, err := models.HasMovies()
	if err != nil {
		log.Printf("Warning: Failed to check if movies exist: %v", err)
	}
	
	if hasMovies {
		log.Println("Movies already exist in database, skipping fetch.")
		return
	}

	log.Println("Fetching movies from Vidsrc API...")

	if err := movies.Fetch(); err != nil {
		log.Fatal(err)
	}

	log.Println("All movies processed and inserted.")
}

func fetchShows() {
	hasShows, err := models.HasShows()
	if err != nil {
		log.Printf("Warning: Failed to check if shows exist: %v", err)
	}
	
	if hasShows {
		log.Println("Shows already exist in database, skipping fetch.")
		return
	}

	log.Println("Fetching shows from Vidsrc API...")

	if err := shows.Fetch(); err != nil {
		log.Fatal(err)
	}

	log.Println("All shows processed and inserted.")
}

func fetchData() {
    if err := godotenv.Load(); err != nil {
	log.Println("No .env file found, using default env vars")
    }

    uri := os.Getenv("MONGO_URI")
    db := os.Getenv("MONGO_DB")
    coll := os.Getenv("MONGO_COLLECTION")

    if err := models.InitMongo(uri, db, coll); err != nil {
	log.Fatal(err)
    }

    if err := models.InitShowMongo(uri, db); err != nil {
	log.Fatal(err)
    }

    if err := models.RebuildTextIndex(); err != nil {
	log.Printf("Warning: Failed to create text index: %v", err)
    }

    fetchMovies()
    fetchShows()

    // models.ClearMoviesCollection()
    // models.ClearShowsCollection()
}

func main() {

    fetchData()

    log.Println("Starting API server...")
    api.RunServer()
}

