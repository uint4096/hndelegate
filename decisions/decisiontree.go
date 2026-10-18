package decisions

import (
	"database/sql"
	"fmt"
)

type DecisionRecord struct {
	Id            int
	Type          string
	Name          string
	Instruction   string
	Parent        sql.Null[int]
	Version       int
	MinConfidence sql.Null[float64]
}

type DecisionRecords struct {
	records []DecisionRecord
}

func getDecisions(db *sql.DB) (DecisionRecords, error) {
	rows, err := db.Query("SELECT id, type, name, instruction, parent, min_confidence from Decisions ORDER BY id")
	if err != nil {
		return DecisionRecords{}, fmt.Errorf("fetch decisions. %w", err)
	}
	defer rows.Close()

	records := []DecisionRecord{}

	for rows.Next() {
		record := DecisionRecord{}
		if err := rows.Scan(&record.Id, &record.Type, &record.Name, &record.Instruction, &record.Parent, &record.MinConfidence); err != nil {
			return DecisionRecords{}, fmt.Errorf("scan decisions. %w", err)
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return DecisionRecords{}, fmt.Errorf("iterate decisions: %w", err)
	}

	return DecisionRecords{records}, nil
}

type Decision struct {
	Id            int
	Type          string
	Name          string
	Instruction   string
	Choices       []Choice
	MinConfidence float64
}

type DecisionTree struct {
	Decision
	Children []*DecisionTree
}

func (d DecisionRecords) tree(choices map[int][]Choice) []*DecisionTree {
	children := make(map[int][]DecisionRecord)
	var roots []DecisionRecord

	for _, record := range d.records {
		if !record.Parent.Valid {
			roots = append(roots, record)
		} else {
			children[record.Parent.V] = append(children[record.Parent.V], record)
		}
	}

	var build func(DecisionRecord) *DecisionTree
	build = func(target DecisionRecord) *DecisionTree {
		d := DecisionTree{
			Id:            target.Id,
			Type:          target.Type,
			Name:          target.Name,
			Instruction:   target.Instruction,
			Choices:       choices[target.Id],
			MinConfidence: target.MinConfidence.V,
		}

		for _, record := range children[target.Id] {
			d.Children = append(d.Children, build(record))
		}

		return &d
	}

	var trees []*DecisionTree
	for _, root := range roots {
		trees = append(trees, build(root))
	}

	return trees
}
