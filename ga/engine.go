package ga

import (
	"fmt"
	"marvin/symbreggenalgo/symbolic"
	"sort"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
)

// HistoryEntry records the best individual at a specific generation.
type HistoryEntry struct {
	Generation int
	Loss       float64
	Expression string
}

// Run the core Symbolic Regression execution loop.
func Run(data Dataset, conf Config, alpha *Alphabet, verbose int) (int, *Individual, *symbolic.Tree, []HistoryEntry, [][10]int) {
	// 1. Initialization
	pop := make(Population, conf.PopulationSize)
	var initWg sync.WaitGroup
	for i := range pop {
		initWg.Add(1)
		go func(idx int) {
			defer initWg.Done()
			pop[idx] = &Individual{
				Tree: GenerateTree(conf.MaxDepth, alpha, conf.Params),
			}
		}(i)
	}
	initWg.Wait()

	if verbose >= 2 {
		tabw := table.NewWriter()
		tabw.SetTitle("Initial Population")
		tabw.AppendHeader(table.Row{"Expressions"})
		for _, ind := range pop {
			tree, err := ind.Tree.ToTree()
			if err != nil {
				fmt.Printf("Error converting tree to string: %v\n", err)
				continue
			}
			tabw.AppendRow(table.Row{tree.String()})
		}
		fmt.Println(tabw.Render() + "\n")
	}

	var bestOverall *Individual
	var history []HistoryEntry
	complexityMeasures := make([][10]int, 0, conf.Generations) // track complexity distributions for each generation for analysis

	// 2. Main Generations Loop
	for generation := 0; generation < conf.Generations; generation++ {

		// 3. Evaluation
		var wg sync.WaitGroup
		for _, ind := range pop {
			wg.Add(1)
			go func(individual *Individual) {
				defer wg.Done()
				EvaluateLoss(individual, data, conf, float64(generation)/float64(conf.Generations))
			}(ind)
		}
		wg.Wait()

		// Sort by best instantaneous loss (raw loss + time dependent complexity penalty)
		sort.Slice(pop, func(i, j int) bool {
			return pop[i].LossInst < pop[j].LossInst
		})

		complexity_bins := [10]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		for _, ind := range pop {
			bin := int((ind.Complexity * 0.8) * 10)
			if bin >= 10 {
				bin = 9
			}
			complexity_bins[bin]++
		}
		complexityMeasures = append(complexityMeasures, complexity_bins)

		if verbose >= 2 {
			tabw := table.NewWriter()
			tabw.SetTitle("Current Population")
			tabw.AppendHeader(table.Row{"Expressions", "LossRaw", "Complexity", "LossInst", "LossFinal"})
			for _, ind := range pop {
				tree, err := ind.Tree.ToTree()
				if err != nil {
					fmt.Printf("Error converting tree to string: %v\n", err)
					continue
				}
				tabw.AppendRow(table.Row{tree.String(), fmt.Sprintf("%0.3f", ind.LossRaw), fmt.Sprintf("%0.3f", ind.Complexity), fmt.Sprintf("%0.3f", ind.LossInst), fmt.Sprintf("%0.3f", ind.LossFinal)})
			}
			fmt.Println(tabw.Render() + "\n")
		}

		// Check if we have a new best overall solution
		currentBest := pop[0]
		if bestOverall == nil || currentBest.LossFinal < bestOverall.LossFinal {
			bestOverall = currentBest.Copy()
			tree, err := bestOverall.Tree.ToTree()
			if err != nil {
				fmt.Printf("Error converting tree to string: %v\n", err)
				continue
			}

			history = append(history, HistoryEntry{
				Generation: generation,
				Loss:       bestOverall.LossFinal,
				Expression: bestOverall.Tree.String(),
			})

			if verbose >= 1 {
				fmt.Printf("New best solution in generation %d: Expression = %s\n\n", generation, tree.String())
				if len(bestOverall.Tree) > 30 {
					fmt.Printf("Note: Expression is quite long (%d nodes).\n", len(bestOverall.Tree))
					fmt.Printf("LossRaw = %0.3f, Complexity = %0.3f, LossInst = %0.3f, LossFinal = %0.3f\n\n", bestOverall.LossRaw, bestOverall.Complexity, bestOverall.LossInst, bestOverall.LossFinal)
				}
			}
		}

		// 4. Mating & Evolution
		newPop := make(Population, 0, conf.PopulationSize)

		// Elitism Carry-over
		for i := 0; i < conf.ElitismCount && i < len(pop); i++ {
			newPop = append(newPop, pop[i].Copy())
		}

		// Selection function based on the configured method
		select_individual := func() *Individual {
			var p1 *Individual
			switch conf.UsedSelection {
			case Tournament:
				p1 = TournamentSelection(pop, conf.SelectionParams.(int))
			case WeightedLoss:
				p1 = LossWeightedSelection(pop)
			default:
				p1 = LossWeightedSelection(pop)
			}
			return p1
		}

		// Mutation helper function
		maybe_mutate_and_add := func(ind *Individual) {
			off := ind.Copy()
			if rnd.Float64() < conf.MutationRate {
				Mutate(off, alpha, conf.MaxDepth, conf.Params)
			}
			newPop = append(newPop, off)
		}

		// Fill new population, apply crossover and mutation
		for len(newPop) < conf.PopulationSize {
			p1 := select_individual()
			if rnd.Float64() < conf.CrossoverRate {
				p2 := select_individual()

				off1, off2 := Crossover(p1, p2)

				maybe_mutate_and_add(off1)
				if len(newPop) < conf.PopulationSize {
					maybe_mutate_and_add(off2)
				}
			} else {
				maybe_mutate_and_add(p1)
			}
		}

		pop = newPop
	}
	generations_taken_for_optimal := -1
	if len(history) > 0 {
		generations_taken_for_optimal = history[len(history)-1].Generation
	}
	bestOverallTree, err := bestOverall.Tree.ToTree()
	if err != nil {
		fmt.Printf("Error converting best overall tree to string: %v\n", err)
	}
	if verbose >= 1 {
		fmt.Printf("Best solution found in generation %d with Loss = %g, Expression = %s\n", generations_taken_for_optimal, bestOverall.LossFinal, bestOverallTree.String())
	}
	return generations_taken_for_optimal, bestOverall, bestOverallTree, history, complexityMeasures
}
