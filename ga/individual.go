package ga

import (
	"marvin/symbreggenalgo/symbolic"
	"math"
	"slices"
)

// Individual represents a single solution in the population.
// LossRaw and Complexity are calculated for the entire population first and then normalized to lie around 1.
// LossInst is time dependent and incorporates raw loss and time dependent complexity penalty.
// LossFinal is LossInst evaluated at t=1 (the final generation).
type Individual struct {
	Tree        symbolic.Postfix
	Complexity  float64   // A measure of the expression's complexity (length of postfix) normalized by conf.MaxComplexity
	LossInst    float64   // Instantaneous loss used for comparison between individuals of the same generation. Incorporates speciation and time dependent complexity penalty.
	LossFinal   float64   // Time independent loss used for comparison between individuals of different generations. Does not include speciation penalty but does include final complexity penalty.
	LossRaw     float64   // The raw loss (e.g., MSE) without any complexity penalty normalizated by conf.MaxLossRaw
	predictions []float64 // Store predictions for each data point
	hasNaN      bool      // Flag to indicate if any prediction was NaN, which can be used to assign worst loss without further calculations
	evaluatedOn string    // store hash or string representation of postfix to avoid re-evaluating
}

// Copy creates a deep copy of an individual, useful for preventing reference issues during crossover and mutation.
func (ind *Individual) Copy() *Individual {
	newTree := make(symbolic.Postfix, len(ind.Tree))
	copy(newTree, ind.Tree)
	return &Individual{
		Tree:        newTree,
		Complexity:  ind.Complexity,
		LossInst:    ind.LossInst,
		LossFinal:   ind.LossFinal,
		LossRaw:     ind.LossRaw,
		predictions: slices.Clone(ind.predictions),
		hasNaN:      ind.hasNaN,
		evaluatedOn: ind.evaluatedOn,
	}
}

// checks if the individual's LossRaw is higher than max allowed loss
func (ind *Individual) IsValid(maxLossRaw float64) bool {
	return ind.LossRaw <= maxLossRaw
}

func (ind *Individual) IsEvaluated() bool {
	return ind.evaluatedOn == ind.Tree.String()
}

// Evaluate computes the predictions for the individual's expression on the dataset and checks for NaN values.
// Avoids redundant evaluations by remembering the string representation of the postfix expression it was evaluated on.
// Assumes that the dataset is not changing between evaluations.
func (ind *Individual) Evaluate(data Dataset) {
	if ind.IsEvaluated() {
		return
	}
	ind.evaluatedOn = ind.Tree.String()
	ind.hasNaN = false
	ind.predictions = make([]float64, len(data))
	for i, dp := range data {
		val, err := EvaluatePostfix(ind.Tree, dp.Variables)
		if err != nil || math.IsNaN(val) || math.IsInf(val, 0) {
			val = math.NaN()
			ind.hasNaN = true
		}
		ind.predictions[i] = val
	}
}

func (ind *Individual) GetPredictions() []float64 {
	return ind.predictions
}

// calculate the similarity or distance between two individuals based on their tree structure alone
func (ind *Individual) DistanceTo(other *Individual, conf *Config) (d float64) {
	// d = c1 * (tree depth difference) + c2 * (size difference) + c3 * (number of different tokens)
	d = 0
	t1, _ := ind.Tree.ToTree()
	t2, _ := other.Tree.ToTree()
	d1 := t1.Depth()
	d2 := t2.Depth()
	ddiff := math.Abs(float64(d1 - d2))
	ddiff /= float64(max(d1, d2))
	d += conf.DifferenceMeasure.TreeDepthWeight * ddiff
	l1 := len(ind.Tree)
	l2 := len(other.Tree)
	ldiff := math.Abs(float64(l1 - l2))
	ldiff /= float64(max(l1, l2))
	d += conf.DifferenceMeasure.TreeSizeWeight * ldiff
	d += conf.DifferenceMeasure.TokenDiffWeight * float64(ind.Tree.TokenDiff(other.Tree)) / float64(max(l1, l2))
	return
}

// Calculates mse between predictions of two individuals
func (ind *Individual) MseTo(other *Individual, data Dataset) (mse float64) {
	ind.Evaluate(data)
	other.Evaluate(data)
	//fmt.Printf("Comparing individuals:\n - %s\n - %s\n", ind.Tree.String(), other.Tree.String())
	if ind.hasNaN && other.hasNaN {
		//fmt.Printf("Both individuals have NaN predictions, assigning distance 0\n")
		return 0.0
	}
	if ind.hasNaN || other.hasNaN {
		//fmt.Printf("One individual has NaN predictions, assigning maximum distance\n")
		return math.MaxFloat64
	}
	mse = 0.0
	for i, dp := range data {
		Y1 := ind.predictions[i]
		Y2 := other.predictions[i]

		diff2 := (Y1 - Y2) * (Y1 - Y2)
		for j, dp2 := range data {
			Y1_2 := ind.predictions[j]
			Y2_2 := other.predictions[j]

			d2 := (Y1_2 - Y2_2) * (Y1_2 - Y2_2)
			for k, v := range dp2.Variables {
				d2 += (dp.Variables[k] - v) * (dp.Variables[k] - v)
			}
			if d2 < diff2 {
				diff2 = d2
			}
		}
		mse += diff2
	}
	mse /= float64(len(data))
	//fmt.Printf("Distance between individuals: %f\n", mse)
	return
}

// Population is a collection of Individuals.
type Population []*Individual
