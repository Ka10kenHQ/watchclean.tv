package scraper

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

func newMovieHTMLElementFromString(html string) *colly.HTMLElement {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	selection := doc.Find("div.post.post-t1").First()

	req := &colly.Request{
		URL: &url.URL{
			Scheme: "https",
			Host:   "mykadri.tv",
		},
		Ctx: &colly.Context{}, 
	}

	return &colly.HTMLElement{
		DOM:     selection,
		Request: req,
	}
}

func TestParseMovie(t *testing.T) {
	html := `
	<div class="post post-t1">
		<a class="post-link post-title-primary" href="/movie-link" title="ქართული ფილმი"></a>
		<a class="post-link post-title-secondary" title="English Title (2023)">English Title (2023)</a>
		<div class="yearshort"><span class="left">2023</span></div>
		<div class="post-image-wrapper">
			<img class="post-image" src="/images/movie.jpg"/>
		</div>
	</div>`

	e := newMovieHTMLElementFromString(html)
	parsed := parseMovie(e)

	if parsed.Title != "ქართული ფილმი" {
		t.Errorf("expected title 'ქართული ფილმი', got %q", parsed.Title)
	}
	if parsed.TitleEnglish != "English Title (2023)" {
		t.Errorf("expected english title, got %q", parsed.TitleEnglish)
	}
	if parsed.Year != "2023" {
		t.Errorf("expected year '2023', got %q", parsed.Year)
	}
	if !strings.Contains(parsed.Image, "movie.jpg") {
		t.Errorf("expected image url to contain 'movie.jpg', got %q", parsed.Image)
	}
	expectedLink := "https://mykadri.tv/movie-link"
	if parsed.Link != expectedLink {
		t.Errorf("expected link %q, got %q", expectedLink, parsed.Link)
	}
}

