package content

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	readability "codeberg.org/readeck/go-readability/v2"
)

const UserAgent = "uint4096/hndelegate (+https://github.com/uint4096/hndelegate)"

type Post struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Fetcher struct {
	client *http.Client
}

func NewFetcher() *Fetcher {
	client := http.Client{
		Timeout: 20 * time.Second,
	}

	return &Fetcher{client: &client}
}

func (f *Fetcher) Get(targetURL string) (*Post, error) {
	request, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("content request. url: %s, err: %w", targetURL, err)
	}

	request.Header.Set("User-Agent", UserAgent)
	request.Header.Set("Accept", "text/html")

	resp, err := f.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("content fetch. url: %s, err: %w", targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("content fetch. url %s, status: %s", resp.Request.URL, resp.Status)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		return nil, fmt.Errorf("content type. url %s, content-type: %s", resp.Request.URL, ct)
	}

	article, err := readability.FromReader(io.LimitReader(resp.Body, 1*1024*1024), resp.Request.URL)
	if err != nil {
		return nil, fmt.Errorf("readability init. url: %s, err: %w", resp.Request.URL, err)
	}

	var text strings.Builder
	if err := article.RenderText(&text); err != nil {
		return nil, fmt.Errorf("article render. url: %s, err: %w", resp.Request.URL, err)
	}

	return &Post{Title: article.Title(), Content: text.String()}, nil
}
