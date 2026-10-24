package classifier

import (
	"fmt"
	"subscribehn/decisions"
	"subscribehn/jev"
)

func FormatQuestions(dTree []*decisions.DecisionTree) (map[string]jev.Question, error) {
	questions := make(map[string]jev.Question)

	var parser func(*decisions.DecisionTree) error
	parser = func(d *decisions.DecisionTree) error {
		question := jev.Question{}
		switch d.Type {
		case string(jev.Choice):
			criteria := make(map[string]string)
			for _, c := range d.Choices {
				description := ""
				if c.Description.Valid {
					description = c.Description.V
				}
				criteria[c.Value] = description
			}

			question = jev.Question{Type: jev.Choice, Instructions: d.Instruction, Criteria: criteria}
		case string(jev.Noul):
			criteria := make(map[string]string)
			question = jev.Question{Type: jev.Noul, Instructions: d.Instruction, Criteria: criteria}
		default:
			return fmt.Errorf("tree parser -> unsupported question type. Decision: %v", d)
		}

		questions[d.Name] = question

		return nil
	}

	if err := decisions.WalkBreadthFirst(dTree, parser); err != nil {
		return questions, err
	}

	return questions, nil
}
