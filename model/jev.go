package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hndelegate/content"
	"hndelegate/decisions"
	"io"
	"net/http"
	"os"
	"slices"
	"time"
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

type Jev struct {
	baseURL    string
	apiKey     string
	httpClient http.Client
}

var _ Model[Answer] = (*Jev)(nil)

func NewJev() (*Jev, error) {
	baseURL := os.Getenv("JEV_BASE_URL")
	apiKey := os.Getenv("JEV_API_KEY")

	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("missing credentials")
	}

	client := http.Client{
		Timeout: 60 * time.Second,
	}

	j := Jev{baseURL, apiKey, client}
	return &j, nil
}

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

func (client *Jev) req(b *Request) (Response, error) {
	if len(b.Questions) == 0 {
		return Response{}, fmt.Errorf("no questions")
	}

	for _, q := range b.Questions {
		if !q.Type.Valid() {
			return Response{}, fmt.Errorf("invalid question type")
		}
	}

	body, err := json.Marshal(b)
	if err != nil {
		return Response{}, fmt.Errorf("marshal request: %w", err)
	}

	r := bytes.NewReader(body)
	request, err := http.NewRequest(http.MethodPost, client.baseURL, r)
	if err != nil {
		return Response{}, fmt.Errorf("build request: %w", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.apiKey))
	request.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(request)
	if err != nil {
		return Response{}, fmt.Errorf("model call: %w", err)
	}
	defer resp.Body.Close()

	rBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("model call status: %d. message: %s", resp.StatusCode, rBody)
	}

	var m Response
	if err := json.Unmarshal(rBody, &m); err != nil {
		return Response{}, fmt.Errorf("unmarshalling response: %w", err)
	}

	return m, nil
}

func (client *Jev) Ask(state *content.Post, roots []*decisions.DecisionTree) (map[string]Answer, error) {
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

func (client *Jev) IsAccepted(answer Answer, d *decisions.DecisionTree) bool {
	choices := d.Choices
	minConfidence := d.MinConfidence

	switch a := answer.(type) {
	case ChoiceResponse:
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
