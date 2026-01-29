# AD Free English Movie and TV series Center

> [!WARNING]
> Always use the stable version available on the **main** branch.



**Landing page**  
![landing page](./images/landing.jpg)  

**Series**
![series](./images/series.jpg)

**Each movie**  
![movie](./images/luther.jpg)

---

### Get started

```sh
git clone https://github.com/Ka10ken1/mykadri-scraper
cd mykadri-scraper
```

**Start with Docker:**

```sh
docker compose up --build
```

The application will be available at: http://localhost:8080

> **Note:** The app uses pre-seeded MongoDB data from `mongo-seed/` directory, so no API key is required for basic usage.

**Optional - TMDB API Key:**

If you want to fetch new movies/shows or refresh metadata, create a `.env` file:

```sh
# .env
TMDB_API_KEY=your_tmdb_api_key_here
```

> Get your free TMDB API key from: https://www.themoviedb.org/settings/api

---

### MongoDB Setup (optional without Docker)

If you're not using Docker, make sure you have MongoDB running:

```sh
sudo systemctl start mongod
```

Then run:

```sh
go run ./cmd/main.go
```

---

### Cleanup

```sh
docker compose down -v
```

---

### Notes

- Scraper skips already-inserted movies (based on link)
- Movie page is scraped for a video iframe
- No retries or slowdowns for HTTP 429 to avoid long waits
- Page concurrency is limited to reduce server stress


### Todo
- Remove Colly and use Chromedp browser
- Integrate full text search engine
- Perfect Scraper
