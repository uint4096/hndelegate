package jev

import (
	"fmt"
	"hndelegate/internal/content"
	"hndelegate/internal/decisions"
	"hndelegate/internal/model"
	"slices"
)

var _ model.Model[Answer] = (*Client)(nil)

const NoneChoice = "None"
const NoneDescription = "None of the other options apply. Choose this if no option is clearly the main subject, if the options are only mentioned in passing, or if the content is about something else entirely."

func formatQuestions(dTree []*decisions.DecisionTree) (map[string]Question, error) {
	questions := make(map[string]Question)

	var parser func(*decisions.DecisionTree) error
	parser = func(d *decisions.DecisionTree) error {
		question := Question{}
		switch d.Type {
		case string(Choice):
			criteria := make(map[string]string)
			for _, c := range d.Choices {
				description := ""
				if c.Description.Valid {
					description = c.Description.V
				}
				criteria[c.Value] = description
			}

			criteria[NoneChoice] = NoneDescription

			question = Question{Type: Choice, Instructions: d.Instruction, Criteria: criteria}
		case string(Noul):
			criteria := make(map[string]string)
			question = Question{Type: Noul, Instructions: d.Instruction, Criteria: criteria}
		default:
			return fmt.Errorf("tree parser -> unsupported question type. Decision: %v", d)
		}

		questions[d.Name] = question

		return nil
	}

	if err := decisions.WalkBreadthFirst(dTree, parser); err != nil {
		return nil, err
	}

	return questions, nil
}

func (client *Client) Ask(state *content.Post, roots []*decisions.DecisionTree) (map[string]Answer, error) {
	questions, err := formatQuestions(roots)
	if err != nil {
		return nil, fmt.Errorf("jev questions: %w", err)
	}

	request := Request{
		State:     state,
		Model:     "jev-latest",
		Questions: questions,
	}

	resp, err := client.req(&request)
	if err != nil {
		return nil, fmt.Errorf("jev request: %w", err)
	}

	return resp.Answers, nil
}

func (client *Client) IsAccepted(answer Answer, d *decisions.DecisionTree) bool {
	choices := d.Choices
	minConfidence := d.MinConfidence

	switch a := answer.(type) {
	case ChoiceResponse:
		if a.Choice == NoneChoice {
			return false
		}

		i := slices.IndexFunc(choices, func(c decisions.Choice) bool {
			return c.Value == a.Choice
		})
		if i == -1 {
			return false
		}
		c := choices[i]

		ap, ok := a.Probabilities[a.Choice]
		if !ok {
			return false
		}

		if (c.AcceptableProbability.Valid && ap < c.AcceptableProbability.V) || a.Confidence < minConfidence {
			return false
		}
	case NoulResponse:
		i := slices.IndexFunc(choices, func(c decisions.Choice) bool {
			return c.AcceptableProbability.Valid && c.AcceptableProbability.V > 0
		})
		if i == -1 {
			return false
		}
		n := choices[i]

		if (n.Value == "no" && a.Noul > 1-n.AcceptableProbability.V) ||
			(n.Value == "yes" && a.Noul < n.AcceptableProbability.V) {
			return false
		}
	default:
		return false
	}

	return true
}
