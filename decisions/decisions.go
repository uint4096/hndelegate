package decisions

import (
	"database/sql"
)

func ConstructTree(db *sql.DB) ([]*DecisionTree, error) {
	d, err := getDecisions(db)
	if err != nil {
		return []*DecisionTree{}, err
	}

	_choices, err := getChoices(db)
	if err != nil {
		return []*DecisionTree{}, err
	}

	choices := mapChoicesByDecisions(_choices)

	decisions := d.tree(choices)

	return decisions, nil
}

type ParserFunction = func(*DecisionTree)

func WalkBreadthFirst(tree []*DecisionTree, parser ParserFunction) {
	for _, node := range tree {
		queue := []*DecisionTree{node}

		for i := 0; i < len(queue); i++ {
			decision := queue[i]
			parser(decision)
			queue = append(queue, decision.Children...)
		}
	}
}
