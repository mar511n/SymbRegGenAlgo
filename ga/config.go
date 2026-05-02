package ga

import (
	"marvin/symbreggenalgo/symbolic"
)

type SelectionMethod int

const (
	Tournament SelectionMethod = iota
	WeightedLoss
)

// Config holds the hyperparameters for the Genetic Algorithm.
type Config struct {
	PopulationSize      int              // Number of individuals in the population
	Generations         int              // Number of generations to evolve
	MaxDepth            int              // Maximum depth of an initialized tree
	CrossoverRate       float64          // Probability of crossover
	MutationRate        float64          // Probability of mutation
	MaxLossRaw          float64          // Estimated maximum expected raw loss for normalization
	MaxComplexity       float64          // Estimated maximum expected complexity for normalization
	MinComplexityWeight float64          // Weight at t=0 for complexity in the combined loss [0,1]
	MaxComplexityWeight float64          // Weight at t=1 for complexity in the combined loss [0,1]
	UsedSelection       SelectionMethod  // Method for selecting individuals for reproduction
	SelectionParams     interface{}      // Parameters for the selected selection method
	ElitismCount        int              // Number of best individuals to carry over automatically
	Params              *GeneratorParams // Parameters for tree generation (e.g., operator probabilities)
}

// DefaultConfig provides a reasonable starting point.
func DefaultConfig() Config {
	return Config{
		PopulationSize:      400,
		Generations:         600,
		MaxDepth:            2,
		CrossoverRate:       0.6,
		MutationRate:        0.7,
		MaxLossRaw:          1e6, // This should be set based on the expected scale of the problem
		MaxComplexity:       10,  // This should be set based on the expected complexity of solutions
		MinComplexityWeight: 0.0,
		MaxComplexityWeight: 0.5, // Start with no complexity penalty and increase to a moderate penalty by the end
		UsedSelection:       Tournament,
		SelectionParams:     4, // Tournament size
		ElitismCount:        4,
		Params:              DefaultGeneratorParams(),
	}
}

// Alphabet defines the allowed operators and terminals for random tree building.
type Alphabet struct {
	Variables []string
	BinaryOps []symbolic.BinaryOp
	UnaryOps  []symbolic.UnaryOp

	// Constants logic
	UseConstants bool
	MinConst     float64
	MaxConst     float64
}

// DefaultAlphabet provides standard mathematical operators.
func DefaultAlphabet(variables []string) *Alphabet {
	return &Alphabet{
		Variables:    variables,
		BinaryOps:    []symbolic.BinaryOp{symbolic.Add, symbolic.Sub, symbolic.Mul, symbolic.Div, symbolic.Pow, symbolic.Max, symbolic.Min, symbolic.Mod},
		UnaryOps:     []symbolic.UnaryOp{symbolic.Sin, symbolic.Cos, symbolic.Exp, symbolic.Log, symbolic.Abs, symbolic.Floor, symbolic.Asin, symbolic.Acos, symbolic.Atan},
		UseConstants: true,
		MinConst:     -10.0,
		MaxConst:     10.0,
	}
}
