package main

import (
	"fmt"
	"log"
	"os"
	"subscribehn/classifier"
	"subscribehn/sqlite"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
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

	title := "How I use HTMX with Go"
	content, err := os.ReadFile("article.txt")
	if err != nil {
		log.Fatal(err)
	}

	questions, err := classifier.FormatQuestions(db)
	if err != nil {
		log.Fatal(err)
	}

	accepted, err := classifier.Classify(classifier.Post{Title: title, Content: string(content)}, questions)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Accpted: %t\n", accepted)
}
