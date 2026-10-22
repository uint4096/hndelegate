package jev

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient http.Client
}

func NewClient() (*Client, error) {
	baseURL := os.Getenv("JEV_BASE_URL")
	apiKey := os.Getenv("JEV_API_KEY")

	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("missing credentials")
	}

	client := http.Client{
		Timeout: 60 * time.Second,
	}

	j := Client{baseURL, apiKey, client}
	return &j, nil
}

func (client *Client) req[T any](b *Request[T]) (Response, error) {
	for _, q := range b.Questions {
		if !q.Type.Valid() {
			return Response{}, fmt.Errorf("invalid question type")
		}
	}

	body, err := json.Marshal(b)
	if err != nil {
		return Response{}, fmt.Errorf("marshal request: %w", err)
	}

	r := bytes.NewReader(body)
	request, err := http.NewRequest(http.MethodPost, client.baseURL, r)
	if err != nil {
		return Response{}, fmt.Errorf("build request: %w", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.apiKey))
	request.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(request)
	if err != nil {
		return Response{}, fmt.Errorf("model call: %w", err)
	}
	defer resp.Body.Close()

	rBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("model call status: %d. message: %s", resp.StatusCode, rBody)
	}

	var m Response
	if err := json.Unmarshal(rBody, &m); err != nil {
		return Response{}, fmt.Errorf("unmarshalling response: %w", err)
	}

	return m, nil
}

func (client *Client) Ask[T any](state T, questions map[string]Question) (*Response, error) {
	request := Request[T]{
		State:     state,
		Model:     "jev-latest",
		Questions: questions,
	}

	resp, err := client.req(&request)
	if err != nil {
		return nil, fmt.Errorf("jev request: %w", err)
	}

	return &resp, nil
}
