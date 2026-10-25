package model

import (
	"hndelegate/internal/content"
	"hndelegate/internal/decisions"
)

type Model[A any] interface {
	Ask(state *content.Post, roots []*decisions.DecisionTree) (map[string]A, error)
	IsAccepted(a A, d *decisions.DecisionTree) bool
}
