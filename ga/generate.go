package ga

import (
	"marvin/symbreggenalgo/symbolic"
)

// GeneratorParams defines probabilities and distributions for generating random trees.
type GeneratorParams struct {
	EarlyTerminationProb float64
	BinaryProb           float64
	ConstantProb         float64
	ConstantGenerator    func() float64
	BinaryOpWeights      map[symbolic.BinaryOp]float64
	UnaryOpWeights       map[symbolic.UnaryOp]float64
	PointMutationProb    float64
	SubtreeMutationProb  float64
}

// DefaultGeneratorParams provides default generation probabilities.
func DefaultGeneratorParams() *GeneratorParams {
	return &GeneratorParams{
		EarlyTerminationProb: 0.15,
		BinaryProb:           0.50,
		ConstantProb:         0.50,
		ConstantGenerator:    nil,
		BinaryOpWeights:      make(map[symbolic.BinaryOp]float64),
		UnaryOpWeights:       make(map[symbolic.UnaryOp]float64),
		PointMutationProb:    0.33,
		SubtreeMutationProb:  0.33,
	}
}

// GenerateTree Grow mode generates random trees (subtrees) to populate postfix array.
func GenerateTree(maxDepth int, alphabet *Alphabet, params *GeneratorParams) symbolic.Postfix {
	if params == nil {
		params = DefaultGeneratorParams()
	}
	var p symbolic.Postfix
	generateRecursive(&p, 0, maxDepth, alphabet, params)
	return p.Simplify()
}

func generateRecursive(p *symbolic.Postfix, currentDepth, maxDepth int, alpha *Alphabet, params *GeneratorParams) {
	// If we've reached max depth, force a terminal (variable or constant)
	isTerminal := currentDepth >= maxDepth
	// Otherwise, maybe pick a terminal randomly (Grow method)
	if !isTerminal && currentDepth > 1 && rnd.Float64() < params.EarlyTerminationProb {
		isTerminal = true
	}

	if isTerminal || (len(alpha.BinaryOps) == 0 && len(alpha.UnaryOps) == 0) {
		// Terminal node
		generateTerminal(p, alpha, params)
		return
	}

	// Calculate weights or probabilities between Unary and Binary
	useBinary := true
	if len(alpha.UnaryOps) > 0 && len(alpha.BinaryOps) > 0 {
		useBinary = rnd.Float64() < params.BinaryProb
	} else if len(alpha.BinaryOps) == 0 {
		useBinary = false
	}

	if useBinary {
		// 1. Generate Left Child
		generateRecursive(p, currentDepth+1, maxDepth, alpha, params)
		// 2. Generate Right Child
		generateRecursive(p, currentDepth+1, maxDepth, alpha, params)
		// 3. Append Operator (postfix layout)
		op := selectBinaryOp(alpha.BinaryOps, params.BinaryOpWeights)
		*p = append(*p, symbolic.Token{Type: symbolic.TokenTypeBinary, BinaryOp: op})
	} else {
		// 1. Generate Input Child
		generateRecursive(p, currentDepth+1, maxDepth, alpha, params)
		// 2. Append Operator
		op := selectUnaryOp(alpha.UnaryOps, params.UnaryOpWeights)
		*p = append(*p, symbolic.Token{Type: symbolic.TokenTypeUnary, UnaryOp: op})
	}
}

func generateTerminal(p *symbolic.Postfix, alpha *Alphabet, params *GeneratorParams) {
	useConst := alpha.UseConstants
	if useConst && len(alpha.Variables) > 0 {
		useConst = rnd.Float64() < params.ConstantProb
	} else if len(alpha.Variables) == 0 {
		useConst = true // Force constant if no variables
	}

	if useConst {
		var val float64
		if params.ConstantGenerator != nil {
			val = params.ConstantGenerator()
		} else {
			val = alpha.MinConst + rnd.Float64()*(alpha.MaxConst-alpha.MinConst)
		}
		*p = append(*p, symbolic.Token{Type: symbolic.TokenTypeConstant, Value: val})
	} else {
		varName := alpha.Variables[rnd.Intn(len(alpha.Variables))]
		*p = append(*p, symbolic.Token{Type: symbolic.TokenTypeVariable, Name: varName})
	}
}

func selectBinaryOp(ops []symbolic.BinaryOp, weights map[symbolic.BinaryOp]float64) symbolic.BinaryOp {
	if len(weights) == 0 {
		return ops[rnd.Intn(len(ops))]
	}
	var total float64
	for _, op := range ops {
		if w, ok := weights[op]; ok {
			total += w
		} else {
			total += 1.0
		}
	}
	if total <= 0 {
		return ops[rnd.Intn(len(ops))]
	}
	r := rnd.Float64() * total
	for _, op := range ops {
		w := 1.0
		if wk, ok := weights[op]; ok {
			w = wk
		}
		r -= w
		if r <= 0 {
			return op
		}
	}
	return ops[len(ops)-1]
}

func selectUnaryOp(ops []symbolic.UnaryOp, weights map[symbolic.UnaryOp]float64) symbolic.UnaryOp {
	if len(weights) == 0 {
		return ops[rnd.Intn(len(ops))]
	}
	var total float64
	for _, op := range ops {
		if w, ok := weights[op]; ok {
			total += w
		} else {
			total += 1.0
		}
	}
	if total <= 0 {
		return ops[rnd.Intn(len(ops))]
	}
	r := rnd.Float64() * total
	for _, op := range ops {
		w := 1.0
		if wk, ok := weights[op]; ok {
			w = wk
		}
		r -= w
		if r <= 0 {
			return op
		}
	}
	return ops[len(ops)-1]
}
