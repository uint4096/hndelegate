package decisions

import (
	"database/sql"
	"fmt"
)

type DecisionRecord struct {
	Id          int
	Type        string
	Instruction string
	Parent      sql.Null[int]
	Version     int
}

type DecisionRecords struct {
	records []DecisionRecord
}

func getDecisions(db *sql.DB) (DecisionRecords, error) {
	rows, err := db.Query("SELECT id, type, instruction, parent from Decisions ORDER BY id")
	if err != nil {
		return DecisionRecords{}, fmt.Errorf("fetch decisions. %w", err)
	}
	defer rows.Close()

	records := []DecisionRecord{}

	for rows.Next() {
		record := DecisionRecord{}
		if err := rows.Scan(&record.Id, &record.Type, &record.Instruction, &record.Parent); err != nil {
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
	Id          int
	Type        string
	Instruction string
	Choices     []Choice
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
		d := DecisionTree{Id: target.Id, Type: target.Type, Instruction: target.Instruction, Choices: choices[target.Id]}
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
