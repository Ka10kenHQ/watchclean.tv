package bootstrap

import (
	"log"
	"os"

	"github.com/Ka10kenHQ/watchclean.tv/internal/handlers"
	"github.com/Ka10kenHQ/watchclean.tv/internal/models"
	"github.com/joho/godotenv"
)

func Init() {
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

	handlers.FetchMovies()
	handlers.FetchShows()
}
