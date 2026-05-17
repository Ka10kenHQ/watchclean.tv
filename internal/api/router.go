package api

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

func RunServer() {
	const port = ":8080"
	r := gin.Default()
	r.Use(cacheHeaders())

	r.GET("/api/movies", GetMovies)
	r.GET("/api/movies/:id", GetMovieByID)
	r.GET("/api/movie-images", GetMovieImages)
	r.GET("/api/search", GetMoviesByTitle)
	r.GET("/api/movie/:id", ShowMoviePage)
	r.GET("/movies/tmdb/:tmdb_id", ShowMovieByTmdbID)

	r.GET("/api/shows", GetShows)
	r.GET("/api/shows/:id", GetShowByID)
	r.GET("/api/shows/images", GetShowImages)
	r.GET("/api/shows/search", GetShowsByTitle)
	r.GET("/api/show/:id", ShowShowPage)
	r.GET("/shows/tmdb/:tmdb_id", ShowShowByTmdbID)

	r.Static("/static", "./web")
	r.Static("/images", "./images")

	
	r.GET("/", func(c *gin.Context) {
		c.File("./web/media.html")
	})


	r.GET("/movies", func(c *gin.Context) {
		c.File("./web/movies.html")
	})

	r.GET("/shows", func(c *gin.Context) {
		c.File("./web/shows.html")
	})

	r.LoadHTMLGlob("web/template/*")

	log.Printf("Starting API server on %s\n", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func cacheHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		switch {
		case strings.HasPrefix(path, "/static/"):
			// Keep this modest because asset filenames are not content-hashed yet.
			c.Header("Cache-Control", "public, max-age=3600")
		case strings.HasPrefix(path, "/images/"):
			c.Header("Cache-Control", "public, max-age=86400")
		case path == "/api/movie-images" || path == "/api/shows/images":
			c.Header("Cache-Control", "public, max-age=600, stale-while-revalidate=300")
		case path == "/api/search" || path == "/api/shows/search":
			c.Header("Cache-Control", "public, max-age=60")
		case strings.HasPrefix(path, "/api/"):
			c.Header("Cache-Control", "no-store")
		default:
			c.Header("Cache-Control", "no-cache")
		}

		c.Next()
	}
}
