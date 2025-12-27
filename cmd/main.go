package main

import (
	"log"
	"os"
	"time"

	"github.com/Ka10ken1/mykadri-scraper/internal/api"
	"github.com/Ka10ken1/mykadri-scraper/internal/models"
	"github.com/Ka10ken1/mykadri-scraper/internal/proxy"
	"github.com/Ka10ken1/mykadri-scraper/internal/scraper"
	"github.com/joho/godotenv"
)


func initProxyPool() *proxy.ProxyClient {
    // free proxy servers
    // replaces with what you like
    // or you can create custom proxy
    pool := proxy.NewPool([]string {
	"http://127.0.0.1:5018",
    })

    return proxy.NewProxyClient(pool)
}

func main() {

    go func ()  {
	proxy.NewServer("127.0.0.1:5018").Start()
    }()
    time.Sleep(2 * time.Second)

    proxyClient := initProxyPool()

    if err := godotenv.Load(); err != nil {
	log.Println("No .env file found, using default env vars")
    }

    uri := os.Getenv("MONGO_URI")
    db := os.Getenv("MONGO_DB")
    coll := os.Getenv("MONGO_COLLECTION")

    if err := models.InitMongo(uri, db, coll); err != nil {
	log.Fatal(err)
    }

    if err := models.InitShowMongo(uri, db, "shows"); err != nil {
	log.Fatal(err)
    }


    if err := models.RebuildTextIndex(); err != nil {
	log.Fatalf("Failed to create text index: %v", err)
    }

    movies, err := scraper.ScrapeMovies(proxyClient)
    if err != nil {
	log.Fatal(err)
    }

    if len(movies) > 0 {

	if err := models.InsertMovies(movies); err != nil {
	    log.Fatal(err)
	}

    } else {
	log.Println("No new movies to insert, skipping DB insert.")
    }

    shows, err := scraper.ScrapeShows(proxyClient)
    if err != nil {
	log.Fatal("Show scrape failed:", err)
    }
    if len(shows) > 0 {
	if err := models.InsertShows(shows); err != nil {
	    log.Fatal("Show insert failed:", err)
	}
    } else {
	log.Println("No new shows to insert.")
    }

    api.RunServer()

    // models.ClearMoviesCollection()
    // models.ClearShowsCollection()

}

