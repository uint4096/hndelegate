package classifier

import (
	"database/sql"
	"fmt"
	"hndelegate/internal/content"
	"hndelegate/internal/decisions"
	"hndelegate/internal/model"
	"hndelegate/internal/transport"
	"slices"
	"strings"
)

type HNResult struct {
	ID   int
	Err  error
	Item *content.HNItem
}

func WalkHNItems(
	itemIDs []int,
	fn func(item *content.HNItem, post *content.Post) error,
	skip func(item *content.HNItem) (bool, error),
) []HNResult {
	hnClient := content.NewHNClient()

	result := []HNResult{}
	fetcher := content.NewFetcher()
	ch := make(chan HNResult)
	semaphore := make(chan struct{}, 5)

	for _, id := range itemIDs {
		go func() {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			item, err := hnClient.GetItem(id)
			if err != nil {
				ch <- HNResult{id, err, item}
				return
			}

			s, err := skip(item)
			if err != nil {
				ch <- HNResult{id, err, item}
				return
			}

			if s {
				ch <- HNResult{ID: id, Item: item}
				return
			}

			post, err := fetcher.Get(item.URL)
			if err != nil {
				ch <- HNResult{id, err, item}
				return
			}

			ch <- HNResult{ID: id, Err: fn(item, post), Item: item}
		}()
	}

	for range itemIDs {
		hnRes := <-ch
		result = append(result, hnRes)
	}

	return result
}

func ClassifyAndSend[A any](
	itemIDs []int,
	roots []*decisions.DecisionTree,
	model model.Model[A],
	t transport.Transport,
	db *sql.DB,
) []HNResult {
	if len(itemIDs) == 0 {
		return []HNResult{}
	}

	w := func(item *content.HNItem, post *content.Post) error {
		resp, err := model.Ask(post, roots)
		if err != nil {
			return fmt.Errorf("model ask: %w", err)
		}

		var isAccepted func(nodes []*decisions.DecisionTree) bool
		isAccepted = func(nodes []*decisions.DecisionTree) bool {
			if len(nodes) == 0 {
				return true
			}

			return slices.ContainsFunc(nodes, func(n *decisions.DecisionTree) bool {
				answer, ok := resp[n.Name]
				return ok && model.IsAccepted(answer, n) && isAccepted(n.Children)
			})
		}

		if isAccepted(roots) {
			message := transport.Post{
				Title:         item.Title,
				ArticleURL:    item.URL,
				DiscussionURL: fmt.Sprintf(`https://news.ycombinator.com/item?id=%d`, item.ID),
			}

			if err := t.Send(message); err != nil {
				return fmt.Errorf("transport send: %w", err)
			}
		}

		return nil
	}

	s := func(item *content.HNItem) (bool, error) {
		if item.Type != content.Story || item.URL == "" || strings.HasPrefix(item.Title, "Show HN:") {
			return true, nil
		}

		return content.TitleExists(db, item.Title)
	}

	errors := WalkHNItems(itemIDs, w, s)

	return errors
}
