package decisions

import (
	"database/sql"
	"fmt"
)

type Choice struct {
	Id                    int
	Value                 string
	Decision              int
	AcceptableProbability sql.Null[float64]
}

func getChoices(db *sql.DB) ([]Choice, error) {
	rows, err := db.Query("SELECT id, value, decision, acceptable_probability FROM Choices ORDER BY id")
	if err != nil {
		return []Choice{}, fmt.Errorf("fetch choices. %w", err)
	}
	defer rows.Close()

	choices := []Choice{}
	for rows.Next() {
		choice := Choice{}
		if err := rows.Scan(&choice.Id, &choice.Value, &choice.Decision, &choice.AcceptableProbability); err != nil {
			return []Choice{}, fmt.Errorf("scan choices. %w", err)
		}

		choices = append(choices, choice)
	}

	if err := rows.Err(); err != nil {
		return []Choice{}, fmt.Errorf("choices iteration: %w", err)
	}

	return choices, nil
}

func mapChoicesByDecisions(choices []Choice) map[int][]Choice {
	c := make(map[int][]Choice)
	for _, choice := range choices {
		c[choice.Decision] = append(c[choice.Decision], choice)
	}

	return c
}
