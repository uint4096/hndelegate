package classifier

import (
	"slices"

	"subscribehn/decisions"
	"subscribehn/jev"
)

func isAccepted(roots []*decisions.DecisionTree, answers jev.Answers) bool {
	var accepted func(*decisions.DecisionTree) bool
	accepted = func(t *decisions.DecisionTree) bool {
		answer, ok := answers[t.Name]
		if !ok {
			return false
		}

		switch a := answer.(type) {
		case jev.ChoiceResponse:
			i := slices.IndexFunc(t.Choices, func(c decisions.Choice) bool {
				return c.Value == a.Choice
			})
			if i == -1 {
				return false
			}
			c := t.Choices[i]

			ap, ok := a.Probabilities[a.Choice]
			if !ok {
				return false
			}

			if (c.AcceptableProbability.Valid && ap < c.AcceptableProbability.V) || a.Confidence < t.MinConfidence {
				return false
			}
		case jev.NoulResponse:
			i := slices.IndexFunc(t.Choices, func(c decisions.Choice) bool {
				return c.AcceptableProbability.Valid && c.AcceptableProbability.V > 0
			})
			if i == -1 {
				return false
			}
			n := t.Choices[i]

			if (n.Value == "no" && a.Noul > 1-n.AcceptableProbability.V) ||
				(n.Value == "yes" && a.Noul < n.AcceptableProbability.V) {
				return false
			}
		default:
			return false
		}

		if len(t.Children) == 0 {
			return true
		}

		return slices.ContainsFunc(t.Children, accepted)
	}

	return slices.ContainsFunc(roots, accepted)
}
