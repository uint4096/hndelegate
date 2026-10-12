package main

import (
	"subscribehn/sqlite"
)

func main() {
	db := sqlite.InitDb()
	defer db.Close()
}
