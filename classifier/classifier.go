package classifier

import (
	"fmt"
	"subscribehn/jev"
)

type Post struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func Classify(state Post, questions map[string]jev.Question) (bool, error) {
	if len(questions) == 0 {
		return false, fmt.Errorf("no questions")
	}

	req := jev.Request[Post]{
		State:     state,
		Model:     "jev-latest",
		Questions: questions,
	}

	client, err := jev.NewClient()
	if err != nil {
		return false, fmt.Errorf("jev client init: %w", err)
	}

	resp, err := client.Ask(&req)
	if err != nil {
		return false, fmt.Errorf("jev request: %w", err)
	}

	fmt.Printf("%+v", resp)

	return true, nil
}
