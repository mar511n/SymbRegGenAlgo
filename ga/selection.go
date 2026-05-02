package ga

import "math"

// Assumes LossInst is roughly in [0,1] and lower is better. Selects individuals with probability roughly proportional to (1 - LossInst).
func LossWeightedSelection(pop Population) *Individual {
	if len(pop) == 0 {
		return nil
	}
	candidate := pop[rnd.Intn(len(pop))]
	for range len(pop) / 10 {
		p := 1.0 - math.Exp(-candidate.LossInst/0.42)
		if rnd.Float64() < 1.0-p {
			return candidate
		}
		candidate = pop[rnd.Intn(len(pop))]
	}
	return candidate
}

// TournamentSelection randomly pulls k candidates from population and picks best.
func TournamentSelection(pop Population, k int) *Individual {
	if len(pop) == 0 {
		return nil
	}
	best := pop[rnd.Intn(len(pop))]
	for i := 1; i < k; i++ {
		candidate := pop[rnd.Intn(len(pop))]
		if candidate.LossInst < best.LossInst {
			best = candidate
		}
	}
	return best
}
