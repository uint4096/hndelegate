package classifier

import (
	"subscribehn/jev"
	"subscribehn/transport"
)

func ClassifyAndSend(target string, questions map[string]jev.Question, t transport.Transport) (bool, error) {
	// 	fetcher := content.NewFetcher()
	//
	// 	state, err := fetcher.Get(target)
	// 	if err != nil {
	// 		return false, fmt.Errorf("post fetch %s: %w", target, err)
	// 	}
	//
	// 	if len(questions) == 0 {
	// 		return false, fmt.Errorf("no questions")
	// 	}
	//
	// 	req := jev.Request[*content.Post]{
	// 		State:     state,
	// 		Model:     "jev-latest",
	// 		Questions: questions,
	// 	}
	//
	// 	client, err := jev.NewClient()
	// 	if err != nil {
	// 		return false, fmt.Errorf("jev client init: %w", err)
	// 	}
	//
	// 	resp, err := client.Ask(&req)
	// 	if err != nil {
	// 		return false, fmt.Errorf("jev request: %w", err)
	// 	}
	//
	// 	message := transport.Post{
	// 		Title:         state.Title,
	// 		ArticleURL:    target,
	// 		DiscussionURL: "https://news.ycombinator.com",
	// 	}
	//
	// 	if err := t.Send(message); err != nil {
	// 		return false, fmt.Errorf("transport send: %w", err)
	// 	}
	//
	// 	fmt.Printf("%+v", resp)
	//
	return true, nil
}
