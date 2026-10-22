package main

import (
	"errors"
	"io/fs"
	"log"
	"subscribehn/sqlite"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatal(err)
	}

	db := sqlite.InitDb()
	defer db.Close()

	// dTree, err := decisions.ConstructTree(db)
	//
	//	if err != nil {
	//		log.Fatal()
	//	}
	//
	// var parser func(*decisions.DecisionTree) error
	//
	//	parser = func(tree *decisions.DecisionTree) error {
	//		fmt.Printf("Type: %s, Instructions: %s, Choices: %v \n", tree.Type, tree.Instruction, tree.Choices)
	//
	//		return nil
	//	}
	//
	// decisions.WalkBreadthFirst(dTree, parser)

	//	questions, err := classifier.FormatQuestions(db)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	transport, err := transport.NewTelegram()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	// accepted, err := classifier.ClassifyAndSend(
	//
	//	"https://davidoks.blog/p/why-china-got-rich-and-india-didnt",
	//	questions,
	//	transport,
	//
	// )
	//
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	// fmt.Printf("Accpted: %t\n", accepted)
}
