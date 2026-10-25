package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Telegram struct {
	client      *http.Client
	requestURL  string
	chatID      string
	errorChatID string
}

func NewTelegram() (*Telegram, error) {
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	errorChatID := os.Getenv("TELEGRAM_ERROR_CHAT_ID")
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	baseURL := os.Getenv("TELEGRAM_BASE_URL")

	if chatID == "" || botToken == "" || baseURL == "" || errorChatID == "" {
		return nil, fmt.Errorf("missing telegram creds")
	}

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	requestURL := fmt.Sprintf("%s/bot%s", baseURL, botToken)

	return &Telegram{chatID: chatID, errorChatID: errorChatID, client: &client, requestURL: requestURL}, nil
}

type telegramRequestBody struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

func (t *Telegram) send(chatID string, msg string) error {
	u := fmt.Sprintf("%s/%s", t.requestURL, "sendMessage")
	body := telegramRequestBody{ChatID: chatID, Text: msg, ParseMode: "HTML"}

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

func (t *Telegram) Send(msg Post) error {
	message := fmt.Sprintf(
		`%s`+"\n"+`<a href="%s">Article</a> `+`<a href="%s">Discussion</a>`,
		html.EscapeString(msg.Title),
		html.EscapeString(msg.ArticleURL),
		html.EscapeString(msg.DiscussionURL),
	)

	return t.send(t.chatID, message)
}

func (t *Telegram) Alert(msg string) {
	m := fmt.Sprintf(`%s`, html.EscapeString(msg[:min(len(msg), 3500)]))
	if err := t.send(t.errorChatID, m); err != nil {
		log.Printf("failed alert. err: %v. msg: %s\n", err, msg)
	}
}
