package ga

import (
	"marvin/symbreggenalgo/symbolic"
)

type SelectionMethod int

const (
	Tournament SelectionMethod = iota
	WeightedLoss
)

type DifferenceMeasureConfig struct {
	TreeSizeWeight  float64
	TokenDiffWeight float64
}

func DefaultDifferenceMeasureConfig() DifferenceMeasureConfig {
	return DifferenceMeasureConfig{
		TreeSizeWeight:  0.5,
		TokenDiffWeight: 0.5,
	}
}

// Config holds the hyperparameters for the Genetic Algorithm.
type Config struct {
	PopulationSize         int                     // Number of individuals in the population
	Generations            int                     // Number of generations to evolve
	MaxDepth               int                     // Maximum depth of an initialized tree
	CrossoverRate          float64                 // Probability of crossover
	MutationRate           float64                 // Probability of mutation
	MaxLossRaw             float64                 // Estimated maximum expected raw loss for normalization if negative, guess from initial population
	MaxComplexity          float64                 // Estimated maximum expected complexity for normalization
	MinComplexityWeight    float64                 // Weight at t=0 for complexity in the combined loss [0,1]
	MaxComplexityWeight    float64                 // Weight at t=1 for complexity in the combined loss [0,1]
	UsedSelection          SelectionMethod         // Method for selecting individuals for reproduction
	SelectionParams        any                     // Parameters for the selected selection method
	GlobalElitismCount     int                     // Number of best individuals to carry over automatically
	SpeciesElites          int                     // Number of elites selected from the top species
	TopElites              int                     // Number of elites selected from the whole population
	CompatibilityThreshold float64                 // Threshold for speciation (to use no speciation, set to a very high value)
	InterSpeciesMatingRate float64                 // Probability of mating between species (0 means no inter-species mating, 1 means random mating)
	MaxValidLossRaw        float64                 // Any individual with LossRaw above this value is considered invalid
	Params                 *GeneratorParams        // Parameters for tree generation (e.g., operator probabilities)
	DifferenceMeasure      DifferenceMeasureConfig // Weights for different components of the distance measure used in speciation
}

// DefaultConfig provides a reasonable starting point.
func DefaultConfig() *Config {
	return &Config{
		PopulationSize:         400,
		Generations:            600,
		MaxDepth:               1,
		CrossoverRate:          0.6,
		MutationRate:           0.7,
		MaxLossRaw:             5e-2, // This should be set based on the expected scale of the problem
		MaxComplexity:          10,   // This should be set based on the expected complexity of solutions
		MinComplexityWeight:    0.0,
		MaxComplexityWeight:    0.5, // Start with no complexity penalty and increase to a moderate penalty by the end
		UsedSelection:          Tournament,
		SelectionParams:        4, // Tournament size
		GlobalElitismCount:     4,
		SpeciesElites:          1,
		TopElites:              1,
		CompatibilityThreshold: 1.0,
		InterSpeciesMatingRate: 0.7,
		MaxValidLossRaw:        1e300,
		Params:                 DefaultGeneratorParams(),
		DifferenceMeasure:      DefaultDifferenceMeasureConfig(),
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
