package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Telegram struct {
	client     *http.Client
	requestURL string
	chatID     string
}

func NewTelegram() (*Telegram, error) {
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	baseURL := os.Getenv("TELEGRAM_BASE_URL")

	if chatID == "" || botToken == "" || baseURL == "" {
		return nil, fmt.Errorf("missing telegram creds")
	}

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	requestURL := fmt.Sprintf("%s/bot%s", baseURL, botToken)

	return &Telegram{chatID: chatID, client: &client, requestURL: requestURL}, nil
}

type telegramRequestBody struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

func (t *Telegram) Send(msg Post) error {
	u := fmt.Sprintf("%s/%s", t.requestURL, "sendMessage")
	message := fmt.Sprintf(`%s`+"\n"+`<a href="%s">Article</a>  `+`<a href="%s">Discussion</a>`, html.EscapeString(msg.Title), html.EscapeString(msg.ArticleURL), html.EscapeString(msg.DiscussionURL))
	body := telegramRequestBody{ChatID: t.chatID, Text: message, ParseMode: "HTML"}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("telegram marshal: %w", err)
	}

	r := bytes.NewReader(b)

	request, err := http.NewRequest(http.MethodPost, u, r)
	if err != nil {
		return fmt.Errorf("telegram body: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(request)
	if err != nil {
		e, ok := errors.AsType[*url.Error](err)
		if ok {
			err = e.Err
		}
		return fmt.Errorf("telegram request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram request: %s", resp.Status)
	}

	return nil
}

func (t *Telegram) Alert(msg string) {
	fmt.Println(msg)
}
