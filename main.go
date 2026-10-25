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

	failedWriteCount := 0
	failedHNCount := 0

	for {
		dTree, err := decisions.ConstructTree(db)
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

		isSeedMode := os.Getenv("SEED_MODE")
		a := []content.Article{}

		if isSeedMode == "" {
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
		} else {
			for _, id := range itemIDs {
				a = append(a, content.Article{
					HNID: id,
				})
			}
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

		if isSeedMode != "" {
			log.Println("seeded article IDs. exiting")
			break
		}

		time.Sleep(120 * time.Second)
	}
}
