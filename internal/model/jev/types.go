package jev

import (
	"encoding/json"
	"fmt"
	"hndelegate/internal/content"
)

type QuestionType string

const (
	Choice QuestionType = "choice"
	Noul   QuestionType = "noul"
	Score  QuestionType = "score"
)

func (q QuestionType) Valid() bool {
	return q == Choice || q == Noul || q == Score
}

type Question struct {
	Type         QuestionType      `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria"`
}

type Request struct {
	State     *content.Post       `json:"state"`
	Model     string              `json:"model"`
	Questions map[string]Question `json:"questions"`
}

type BaseResponse struct {
	Type QuestionType `json:"type"`
}

type NoulResponse struct {
	BaseResponse
	Noul float64 `json:"noul"`
}

type ChoiceResponse struct {
	BaseResponse
	Choice        string             `json:"choice"`
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    float64            `json:"confidence"`
}

type Answer interface {
	isAnswer()
}

func (BaseResponse) isAnswer() {}

type Answers map[string]Answer
type Response struct {
	Answers Answers `json:"answers"`
}

func (r *Answers) UnmarshalJSON(b []byte) error {
	var ans map[string]json.RawMessage
	if err := json.Unmarshal(b, &ans); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	out := make(Answers)
	for key, msg := range ans {
		var a BaseResponse
		if err := json.Unmarshal(msg, &a); err != nil {
			return fmt.Errorf("unmarshall type: %w", err)
		}

		switch a.Type {
		case Choice:
			{
				var c ChoiceResponse
				if err := json.Unmarshal(msg, &c); err != nil {
					return fmt.Errorf("unmarshall choice: %w", err)
				}
				out[key] = c
			}
		case Noul:
			{
				var n NoulResponse
				if err := json.Unmarshal(msg, &n); err != nil {
					return fmt.Errorf("unmarshall noul: %w", err)
				}
				out[key] = n
			}
		default:
			return fmt.Errorf("Invalid type %s", a.Type)
		}
	}

	*r = out
	return nil
}
