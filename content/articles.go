package content

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Article struct {
	ID   int
	HNID int
}

func getArticles(db *sql.DB) ([]Article, error) {
	rows, err := db.Query("SELECT id, hn_id FROM Articles WHERE expires_at > unixepoch() ORDER BY id LIMIT 500")
	if err != nil {
		return nil, fmt.Errorf("articles fetch: %w", err)
	}
	defer rows.Close()

	articles := []Article{}
	for rows.Next() {
		article := Article{}
		if err := rows.Scan(&article.ID, &article.HNID); err != nil {
			return nil, fmt.Errorf("articles scan: %w", err)
		}

		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("article rows: %w", err)
	}

	return articles, nil
}

func createArticles(db *sql.DB, articleIds []int) error {
	if len(articleIds) == 0 {
		return nil
	}

	t := time.Now().Add(24 * 30 * 6 * time.Hour).Unix()
	rows := strings.Builder{}
	for i, article := range articleIds {
		row := fmt.Sprintf("(%d,%d)", article, t)
		rows.WriteString(row)

		if i != len(articleIds)-1 {
			rows.WriteString(",")
		}
	}

	if _, err := db.Exec(fmt.Sprintf("INSERT INTO articles (hn_id, expires_at) VALUES %s ON CONFLICT(hn_id) DO NOTHING", rows.String())); err != nil {
		return fmt.Errorf("articles insert: %w", err)
	}

	return nil
}
