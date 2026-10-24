package classifier

import (
	"fmt"
	"hndelegate/content"
	"hndelegate/decisions"
	"hndelegate/model"
	"hndelegate/transport"
	"slices"
)

type HNError struct {
	ID   int
	Err  error
	Item *content.HNItem
}

func WalkHNItems(itemIDs []int, fn func(item *content.HNItem, post *content.Post) error) []HNError {
	hnClient := content.NewHNClient()

	errors := []HNError{}
	fetcher := content.NewFetcher()
	ch := make(chan HNError)
	semaphore := make(chan struct{}, 5)

	for _, id := range itemIDs {
		go func() {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			item, err := hnClient.GetItem(id)
			if err != nil {
				ch <- HNError{id, err, item}
				return
			}

			if item.Type != content.Story || item.URL == "" {
				ch <- HNError{ID: id, Item: item}
				return
			}

			post, err := fetcher.Get(item.URL)
			if err != nil {
				ch <- HNError{id, err, item}
				return
			}

			ch <- HNError{ID: id, Err: fn(item, post), Item: item}
		}()
	}

	for range itemIDs {
		hnErr := <-ch

		if hnErr.Err != nil {
			errors = append(errors, hnErr)
		}
	}

	return errors
}

func ClassifyAndSend[A any](
	itemIDs []int,
	roots []*decisions.DecisionTree,
	model model.Model[A],
	t transport.Transport,
) []HNError {
	if len(itemIDs) == 0 {
		return []HNError{}
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

	errors := WalkHNItems(itemIDs, w)

	return errors
}
