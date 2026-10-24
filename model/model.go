package model

import (
	"subscribehn/content"
	"subscribehn/decisions"
)

type Model[A any] interface {
	Ask(state *content.Post, roots []*decisions.DecisionTree) (map[string]A, error)
	IsAccepted(a A, d *decisions.DecisionTree) bool
}
