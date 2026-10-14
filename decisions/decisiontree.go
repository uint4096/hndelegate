package decisions

import "database/sql"

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

type Decision struct {
	Id          int
	Type        string
	Instruction string
	Children    []*Decision
}

func (d DecisionRecords) DecisionTrees() []*Decision {
	children := make(map[int][]DecisionRecord)
	var roots []DecisionRecord

	for _, record := range d.records {
		if !record.Parent.Valid {
			roots = append(roots, record)
		} else {
			children[record.Parent.V] = append(children[record.Parent.V], record)
		}
	}

	var build func(DecisionRecord) *Decision
	build = func(target DecisionRecord) *Decision {
		d := Decision{Id: target.Id, Type: target.Type, Instruction: target.Instruction}
		for _, record := range children[target.Id] {
			d.Children = append(d.Children, build(record))
		}

		return &d
	}

	var trees []*Decision
	for _, root := range roots {
		trees = append(trees, build(root))
	}

	return trees
}
