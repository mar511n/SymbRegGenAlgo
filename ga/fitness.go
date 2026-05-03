package ga

import (
	"fmt"
	"marvin/symbreggenalgo/symbolic"
	"math"
)

type DataPoint struct {
	Variables map[string]float64
	Target    float64
}

type Dataset []DataPoint

func EvaluateLoss(ind *Individual, data Dataset, conf *Config, time float64, rel_species_size float64) {
	EvaluateLossRaw(ind, data, conf)
	EvaluateComplexity(ind, conf)
	ind.LossRaw = ind.LossRaw / conf.MaxLossRaw
	ind.Complexity = (ind.Complexity - 1) / (conf.MaxComplexity - 1)
	if ind.Complexity > 1 {
		ind.Complexity = math.Exp(ind.Complexity - 1)
	}
	if math.IsNaN(ind.LossRaw) || math.IsInf(ind.LossRaw, 0) {
		ind.LossRaw = math.MaxFloat64
	}
	if math.IsNaN(ind.Complexity) || math.IsInf(ind.Complexity, 0) {
		ind.Complexity = math.MaxFloat64
	}
	//ind.LossRaw = 1 - math.Exp(-ind.LossRaw/conf.MaxLossRaw)
	//ind.Complexity = 1 - math.Exp(-(ind.Complexity-1)/(conf.MaxComplexity-1))
	a := conf.MinComplexityWeight
	b := conf.MaxComplexityWeight
	w := a + (b-a)*time
	x := ind.LossRaw
	y := ind.Complexity
	//ind.LossInst = (1-w)*x + w*y
	//ind.LossFinal = (1-b)*x + b*y
	ind.LossInst = (1 + w*(y-1)) * (x - w*x + w*y)
	ind.LossFinal = (1 + b*(y-1)) * (x - b*x + b*y)
	ind.LossInst *= rel_species_size
	if math.IsNaN(ind.LossInst) || math.IsInf(ind.LossInst, 0) {
		ind.LossInst = math.MaxFloat64
	}
	if math.IsNaN(ind.LossFinal) || math.IsInf(ind.LossFinal, 0) {
		ind.LossFinal = math.MaxFloat64
	}
	if ind.LossInst == 0 {
		ind.LossInst = 1e-14
	}
}

func EvaluateLossRaw(ind *Individual, data Dataset, conf *Config) {
	ind.Evaluate(data)
	if ind.hasNaN {
		ind.LossRaw = math.MaxFloat64
		return
	}
	mse := 0.0
	for i, dp := range data {
		val := ind.predictions[i]

		diff2 := (val - dp.Target) * (val - dp.Target)
		for _, dp2 := range data {

			d2 := (val - dp2.Target) * (val - dp2.Target)
			for k, v := range dp2.Variables {
				d2 += (dp.Variables[k] - v) * (dp.Variables[k] - v)
			}
			if d2 < diff2 {
				diff2 = d2
			}
		}
		mse += diff2
	}
	ind.LossRaw = mse / float64(len(data))

	// Apply penalty for large expressions
	// treeSize := len(ind.Tree)
	// penaltyf := float64(treeSize) / conf.PenaltySize
	// ind.Loss = mse * (1.0 + penaltyf)
	// penaltyf *= 1.0 - math.Exp(-3*progression)
	// ind.CurrentLoss = mse * (1.0 + penaltyf)

	if math.IsNaN(ind.LossRaw) || math.IsInf(ind.LossRaw, 0) {
		ind.LossRaw = math.MaxFloat64
	}
}

func EvaluateComplexity(ind *Individual, conf *Config) {
	ind.Complexity = float64(len(ind.Tree))
}

// EvaluatePostfix directly computes a postfix expression without allocations other than a stack. TODO: create a more efficient evaluator that operates on a slice of input variables
func EvaluatePostfix(p symbolic.Postfix, vars map[string]float64) (float64, error) {
	stack := make([]float64, 0, len(p))
	for _, token := range p {
		switch token.Type {
		case symbolic.TokenTypeConstant:
			stack = append(stack, token.Value)
		case symbolic.TokenTypeVariable:
			stack = append(stack, vars[token.Name])
		case symbolic.TokenTypeUnary:
			if len(stack) < 1 {
				return 0, fmt.Errorf("stack underflow on unary")
			}
			val := stack[len(stack)-1]
			stack[len(stack)-1] = symbolic.BasicUnaryOps[token.UnaryOp](val)
		case symbolic.TokenTypeBinary:
			if len(stack) < 2 {
				return 0, fmt.Errorf("stack underflow on binary")
			}
			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, symbolic.BasicBinaryOps[token.BinaryOp](left, right))
		}
	}
	if len(stack) != 1 {
		return 0, fmt.Errorf("invalid stack state: %d items left", len(stack))
	}
	return stack[0], nil
}
