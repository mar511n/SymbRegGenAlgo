package ga

import (
	"marvin/symbreggenalgo/symbolic"
)

// Individual represents a single solution in the population.
// LossRaw and Complexity are calculated for the entire population first and then normalized to lie around 1.
// LossInst is time dependent and incorporates raw loss and time dependent complexity penalty.
// LossFinal is LossInst evaluated at t=1 (the final generation).
type Individual struct {
	Tree       symbolic.Postfix
	Complexity float64 // A measure of the expression's complexity (length of postfix)
	LossInst   float64 // Instantaneous loss used for comparison between individuals of the same generation
	LossFinal  float64 // Time independent loss used for comparison between individuals of different generations
	LossRaw    float64 // The raw loss (e.g., MSE) without any complexity penalty
}

// Copy creates a deep copy of an individual, useful for preventing reference issues during crossover and mutation.
func (ind *Individual) Copy() *Individual {
	newTree := make(symbolic.Postfix, len(ind.Tree))
	copy(newTree, ind.Tree)
	return &Individual{
		Tree:       newTree,
		Complexity: ind.Complexity,
		LossInst:   ind.LossInst,
		LossFinal:  ind.LossFinal,
		LossRaw:    ind.LossRaw,
	}
}

// Population is a collection of Individuals.
type Population []*Individual
