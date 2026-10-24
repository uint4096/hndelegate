package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"subscribehn/classifier"
	"subscribehn/content"
	"subscribehn/decisions"
	"subscribehn/jev"
	"subscribehn/sqlite"
	"subscribehn/transport"
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

	model, err := jev.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	failedWriteCount := 0
	failedHNCount := 0

	for {
		dTree, err := decisions.ConstructTree(db)
		if err != nil {
			transport.Alert(err.Error())
			log.Fatal(err)
		}

		questions, err := classifier.FormatQuestions(dTree)
		if err != nil {
			transport.Alert(err.Error())
			log.Fatal(err)
		}

		itemIDs, err := hn.GetNewItemIds(db)
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

		hnErrs := classifier.ClassifyAndSend(
			itemIDs,
			dTree,
			questions,
			model,
			transport,
		)

		for _, hnErr := range hnErrs {
			message := fmt.Sprintf("Err: %v, ID: %d", hnErr.Err, hnErr.ID)
			if hnErr.Item != nil {
				message = fmt.Sprintf("Err: %v, ID: %d, url: %s", hnErr.Err, hnErr.ID, hnErr.Item.URL)
			}

			transport.Alert(message)
		}

		if err := content.CreateArticles(db, itemIDs); err != nil {
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
