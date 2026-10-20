package content

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const HNBaseURL = "https://hacker-news.firebaseio.com/v0"

type StoryIDs []int

type HNClient struct {
	client  *http.Client
	baseURL string
}

func newHNClient() *HNClient {
	client := HNClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: HNBaseURL,
	}

	return &client
}

func (hn *HNClient) get[T any](path string) (*T, error) {
	reqURL, err := url.JoinPath(hn.baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("hn url join: %w", err)
	}

	resp, err := hn.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("hn fetch. url: %s. err: %w", reqURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hn fetch. url: %s, err: %s", reqURL, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("hn response. url: %s, err: %w", reqURL, err)
	}

	var res T
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("hn unmarshal: %w", err)
	}

	return &res, nil
}

func (hn *HNClient) getStoryIDs() (StoryIDs, error) {
	storyIDs, err := hn.get[StoryIDs]("newstories.json")
	if err != nil {
		return nil, err
	}

	return *storyIDs, nil
}

type ItemType string

const (
	Job     ItemType = "job"
	Story   ItemType = "story"
	Comment ItemType = "comment"
	Poll    ItemType = "poll"
	PollOpt ItemType = "pollopt"
)

type HNItem struct {
	By    string   `json:"by"`
	ID    int      `json:"id"`
	Type  ItemType `json:"type"`
	Title string   `json:"title"`
	URL   string   `json:"url"`
}

func (hn *HNClient) getItem(id int) (*HNItem, error) {
	return hn.get[HNItem](fmt.Sprintf("item/%d.json", id))
}

func WithHNStory(db *sql.DB, fn func(*HNItem, string) error) error {
	hnClient := newHNClient()
	storyIDs, err := hnClient.getStoryIDs()
	if err != nil {
		return err
	}

	savedIDs, err := getArticles(db)
	if err != nil {
		return err
	}

	filteredIDs := []int{}
	for _, id := range storyIDs {
		if _, ok := savedIDs[id]; !ok {
			filteredIDs = append(filteredIDs, id)
		}
	}

	fetcher := newFetcher()

	parser := func(id int) error {
		story, err := hnClient.getItem(id)
		if err != nil {
			return err
		}

		article, err := fetcher.get(story.URL)
		if err != nil {
			return err
		}

		return fn(story, article.Content)
	}

	type result struct {
		id  int
		err error
	}

	ch := make(chan result)
	sem := make(chan struct{}, 5)
	for _, id := range filteredIDs {
		go func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			err := parser(id)
			ch <- result{id, err}
		}()
	}

	for range filteredIDs {
		res := <-ch

		if res.err != nil {
			fmt.Printf("hn parser. id: %d. err: %v\n", res.id, res.err)
		}
	}

	if err := createArticles(db, filteredIDs); err != nil {
		return err
	}

	return nil
}
