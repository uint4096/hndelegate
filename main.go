package main

import (
	"fmt"
	"log"
	"subscribehn/decisions"
	"subscribehn/sqlite"
)

func main() {
	db := sqlite.InitDb()
	defer db.Close()

	dTree, err := decisions.ConstructTree(db)
	if err != nil {
		log.Fatal()
	}

	var parser func(*decisions.DecisionTree)
	parser = func(tree *decisions.DecisionTree) {
		fmt.Printf("Type: %s, Instructions: %s, Choices: %v \n", tree.Type, tree.Instruction, tree.Choices)
	}

	decisions.WalkBreadthFirst(dTree, parser)
}
