package scraper

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

func newHTMLElementFromString(html string) *colly.HTMLElement {
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

func TestParseShow(t *testing.T) {
	html := `
	<div class="post post-t1">
		<a class="post-link post-title-primary" href="/show-link" title="ქართული სახელი"></a>
		<a class="post-link post-title-secondary" title="English Title (2022)">English Title (2022)</a>
		<div class="yearshort"><span class="left">2022</span></div>
		<div class="post-image-wrapper">
			<img class="post-image" src="/images/show.jpg"/>
		</div>
	</div>`

	e := newHTMLElementFromString(html)
	parsed := parseShow(e)

	if parsed.Title != "ქართული სახელი" {
		t.Errorf("expected title 'ქართული სახელი', got %q", parsed.Title)
	}

	if parsed.TitleEnglish != "English Title (2022)" {
		t.Errorf("expected english title, got %q", parsed.TitleEnglish)
	}

	if parsed.Year != "2022" {
		t.Errorf("expected year '2022', got %q", parsed.Year)
	}

	if !strings.Contains(parsed.Image, "show.jpg") {
		t.Errorf("expected image url to contain 'show.jpg', got %q", parsed.Image)
	}

	expectedLink := "https://mykadri.tv/show-link"
	if parsed.Link != expectedLink {
		t.Errorf("expected link '/show-link', got %q", parsed.Link)
	}
}

