package main

import (
	"errors"
	"fmt"
	"hndelegate/internal/classifier"
	"hndelegate/internal/content"
	"hndelegate/internal/decisions"
	"hndelegate/internal/model/jev"
	"hndelegate/internal/sqlite"
	"hndelegate/internal/transport"
	"io/fs"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatal(err)
	}

	db := sqlite.InitDb()
	defer db.Close()

	transport, err := transport.NewTelegram()
	if err != nil {
		log.Fatal(err)
	}

	hn := content.NewHNClient()

	jev, err := jev.New()
	if err != nil {
		log.Fatal(err)
	}

	maxCount, err := strconv.Atoi(os.Getenv("MAX_ITEMS_PER_ROUND"))
	if err != nil || maxCount <= 0 {
		maxCount = 20
	}

	failedWriteCount := 0
	failedHNCount := 0

	for {
		dTree, err := decisions.ConstructTree(db)
		if err != nil {
			transport.Alert(err.Error())
			log.Fatal(err)
		}

		ids, err := hn.GetNewItemIds(db)
		if err != nil {
			log.Println(err)

			failedHNCount += 1
			if failedHNCount >= 3 {
				transport.Alert(fmt.Sprintf("error: %s. third consecutive HN item IDs call failed. crashing...", err))
				log.Fatal(err)
			}

			time.Sleep(120 * time.Second)
			continue
		}
		failedHNCount = 0

		itemIDs := ids[:min(maxCount, len(ids))]
		if skipped := len(ids) - len(itemIDs); skipped > 0 {
			log.Printf("skipped %d stories", skipped)
		}

		a := []content.Article{}

		result := classifier.ClassifyAndSend(
			itemIDs,
			dTree,
			jev,
			transport,
			db,
		)

		for _, res := range result {
			title := ""
			if res.Item != nil {
				title = res.Item.Title
			}

			a = append(a, content.Article{
				HNID:  res.ID,
				Title: title,
			})

			if res.Err == nil {
				continue
			}

			message := fmt.Sprintf("Err: %v, ID: %d", res.Err, res.ID)
			if res.Item != nil {
				message = fmt.Sprintf("Err: %v, ID: %d, url: %s", res.Err, res.ID, res.Item.URL)
			}

			log.Printf("hn err: %v", res)
			transport.Alert(message)
		}

		for _, id := range ids[len(itemIDs):] {
			a = append(a, content.Article{HNID: id})
		}

		if err := content.CreateArticles(db, a); err != nil {
			transport.Alert(fmt.Sprintf("articles sqlite write: %v", err))

			failedWriteCount += 1
			if failedWriteCount > 2 {
				msg := "articles write failure threshold reached. crashing"
				transport.Alert(msg)
				log.Fatal(msg)
			}
		} else {
			failedWriteCount = 0
		}

		time.Sleep(120 * time.Second)
	}
}
