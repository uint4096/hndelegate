package content

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Article struct {
	ID    int
	HNID  int
	Title string
}

func GetArticles(db *sql.DB) (map[int]Article, error) {
	rows, err := db.Query("SELECT id, hn_id FROM Articles WHERE expires_at > unixepoch() ORDER BY hn_id desc LIMIT 500")
	if err != nil {
		return nil, fmt.Errorf("articles fetch: %w", err)
	}
	defer rows.Close()

	articles := map[int]Article{}
	for rows.Next() {
		article := Article{}
		if err := rows.Scan(&article.ID, &article.HNID); err != nil {
			return nil, fmt.Errorf("articles scan: %w", err)
		}

		articles[article.HNID] = article
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("article rows: %w", err)
	}

	return articles, nil
}

func TitleExists(db *sql.DB, title string) (bool, error) {
	rows, err := db.Query("SELECT id FROM Articles WHERE title = ? LIMIT 1", title)
	if err != nil {
		return false, fmt.Errorf("articles title check: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		return true, nil
	}

	return false, nil
}

func CreateArticles(db *sql.DB, articles []Article) error {
	if len(articles) == 0 {
		return nil
	}

	t := time.Now().Add(24 * 30 * 6 * time.Hour).Unix()
	rows := make([]string, 0, len(articles))
	args := make([]any, 0, len(articles)*3)

	for _, article := range articles {
		rows = append(rows, "(?, ?, ?)")
		args = append(args, article.HNID, article.Title, t)
	}

	query := "INSERT INTO articles (hn_id, title, expires_at) VALUES " + strings.Join(rows, ",") + " ON CONFLICT(hn_id) DO NOTHING"
	if _, err := db.Exec(query, args...); err != nil {
		return fmt.Errorf("articles insert: %w", err)
	}

	return nil
}
