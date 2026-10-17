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

type ParserFunction = func(*DecisionTree) error

func WalkBreadthFirst(tree []*DecisionTree, parser ParserFunction) error {
	for _, node := range tree {
		queue := []*DecisionTree{node}

		for i := 0; i < len(queue); i++ {
			decision := queue[i]
			if err := parser(decision); err != nil {
				return err
			}

			queue = append(queue, decision.Children...)
		}
	}

	return nil
}
