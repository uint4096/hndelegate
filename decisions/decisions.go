package decisions

import (
	"database/sql"
)

func ConstructDecisionTree(db *sql.DB) ([]*DecisionTree, error) {
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

func ParseDecisionTree(tree []*DecisionTree, parser ParserFunction) {
	for _, decision := range tree {
		parser(decision)
		ParseDecisionTree(decision.Children, parser)
	}
}
