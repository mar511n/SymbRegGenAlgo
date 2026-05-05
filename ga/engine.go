package ga

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"sync"

	"github.com/jedib0t/go-pretty/progress"
	"github.com/jedib0t/go-pretty/v6/table"
)

// TODO: benchmark & parallelize
// Run the core Symbolic Regression execution loop.
func Run(data Dataset, conf *Config, alpha *Alphabet, verbose int, target_num_history_events int) (history *EvolutionHistory) {
	history_dump_step := conf.Generations / target_num_history_events
	history = &EvolutionHistory{
		BestOverall:     nil,
		BestOverallTree: nil,
		HallOfFame:      make(map[int]*Individual),
		Complexity:      make(map[int][10]int),
		Species:         make(map[int][]SpeciesEntry),
	}

	// Selection function based on the configured method
	select_individual := func(population Population) *Individual {
		var p1 *Individual
		switch conf.UsedSelection {
		case Tournament:
			p1 = TournamentSelection(population, conf.SelectionParams.(int))
		case WeightedLoss:
			p1 = LossWeightedSelection(population)
		default:
			p1 = LossWeightedSelection(population)
		}
		return p1
	}

	// Population Initialization
	pop := make(Population, conf.PopulationSize)
	var species map[int]Population // map from species ID to list of individuals in that species
	species_representatives := make(map[int]*Individual)
	species_id_counter := 0
	//var initWg sync.WaitGroup
	for i := range pop {
		//initWg.Add(1)
		//go
		func(idx int) {
			//defer initWg.Done()
			pop[idx] = &Individual{
				Tree: GenerateTree(conf.MaxDepth, alpha, conf.Params),
			}
		}(i)
	}
	//initWg.Wait()

	if verbose >= 6 {
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

	// TODO: use higher depth and more individuals for initial guessing
	if conf.MaxLossRaw < 0 {
		fmt.Printf("Guessing max loss from initial population...\n")
		//var wg sync.WaitGroup
		for _, ind := range pop {
			//wg.Add(1)
			//go
			func(individual *Individual) {
				//defer wg.Done()
				EvaluateLossRaw(individual, data, conf)
			}(ind)
		}
		//wg.Wait()
		conf.MaxLossRaw = math.MaxFloat64
		found := false
		for _, ind := range pop {
			if ind.LossRaw < conf.MaxLossRaw {
				conf.MaxLossRaw = ind.LossRaw
				found = true
			}
		}
		if !found {
			fmt.Printf("ERROR: could not estimate MaxLossRaw from initial Population, no finite LossRaw found.\nTry to increase the population size.")
			return
		}
		fmt.Printf("Estimated MaxLossRaw=%0.4e from initial Population.\n", conf.MaxLossRaw)
	}

	pbar := &progress.Tracker{Total: int64(conf.Generations)}
	if verbose < 2 {
		pbar.Reset()
		pwriter := progress.NewWriter()
		pwriter.AppendTracker(pbar)
		pwriter.ShowPercentage(true)
		pwriter.ShowOverallTracker(true)
		go func() {
			pwriter.Render()
		}()
	}

	// 2. Main Generations Loop
	for generation := 0; generation < conf.Generations; generation++ {

		// Evaluate individuals
		var wg sync.WaitGroup
		for _, ind := range pop {
			wg.Add(1)
			go func(individual *Individual) {
				defer wg.Done()
				ind.Evaluate(data)
			}(ind)
		}
		wg.Wait()

		// Speciation
		// Species representatives
		species_representatives = make(map[int]*Individual)
		for speciesID, members := range species {
			species_representatives[speciesID] = members[0]
		}
		species = make(map[int]Population)
		// assign individuals to species based on distance to representative or create new species
		for _, ind := range pop {
			assigned := false
			for speciesID := range species_id_counter {
				repr, ok := species_representatives[speciesID]
				if ok && ind.DistanceTo(repr, conf) < conf.CompatibilityThreshold {
					if _, exists := species[speciesID]; !exists {
						species[speciesID] = Population{}
					}
					species[speciesID] = append(species[speciesID], ind)
					assigned = true
					break
				}
			}
			if !assigned {
				species[species_id_counter] = Population{ind}
				species_representatives[species_id_counter] = ind
				species_id_counter++
			}
		}

		// Find max species size for relative loss calculation
		maxSpeciesSize := 0
		for _, members := range species {
			if len(members) > maxSpeciesSize {
				maxSpeciesSize = len(members)
			}
		}

		// Loss Evaluation
		for _, members := range species {
			for _, ind := range members {
				wg.Add(1)
				go func(individual *Individual, rel_spec_size float64) {
					defer wg.Done()
					EvaluateLoss(individual, data, conf, float64(generation)/float64(conf.Generations), rel_spec_size)
				}(ind, float64(len(members))/float64(maxSpeciesSize))
			}
		}
		wg.Wait()

		// Sort by best instantaneous loss
		sort.Slice(pop, func(i, j int) bool {
			return pop[i].LossFinal < pop[j].LossFinal
		})
		for _, members := range species {
			sort.Slice(members, func(i, j int) bool {
				return members[i].LossFinal < members[j].LossFinal
			})
		}

		// Compute number of allowed offspring per species based on mean LossInst and number of invalid individuals
		species_fitness := make(map[int]float64)
		species_invalid_counts := make(map[int]int)
		total_fitness := 0.0
		total_invalid := 0
		for speciesID, members := range species {
			species_fitness[speciesID] = 0
			species_invalid_counts[speciesID] = 0
			for _, ind := range members {
				if ind.IsValid(conf.MaxValidLossRaw) {
					species_fitness[speciesID] += 1.0 / ind.LossInst
					total_fitness += 1.0 / ind.LossInst
				} else {
					species_invalid_counts[speciesID]++
					total_invalid++
				}
			}
		}
		fitness_sorted_ids := make([]int, 0, len(species))
		for id := range species {
			fitness_sorted_ids = append(fitness_sorted_ids, id)
		}
		sort.Slice(fitness_sorted_ids, func(i, j int) bool {
			return species_fitness[fitness_sorted_ids[i]] > species_fitness[fitness_sorted_ids[j]]
		})
		offspring_counts := make(map[int]int)
		total_offspring := 0
		for speciesID, fitness := range species_fitness {
			if species_invalid_counts[speciesID] >= len(species[speciesID]) {
				offspring_counts[speciesID] = 0
			} else {
				offspring_counts[speciesID] = int(float64(conf.PopulationSize-conf.GlobalElitismCount) * (fitness / total_fitness))
			}
			total_offspring += offspring_counts[speciesID]
		}
		// adjust elitism count based on total offspring to ensure population size remains constant
		elites_count := conf.PopulationSize - total_offspring

		// Dump history for this generation and print if needed
		if generation%history_dump_step == 0 {
			// collect species information
			speciesEntries := make([]SpeciesEntry, 0, len(species))
			for _, speciesID := range fitness_sorted_ids {
				members := species[speciesID]
				entry := SpeciesEntry{
					ID:             speciesID,
					Size:           len(members),
					Representative: slices.Clone(species_representatives[speciesID].Tree),
					LossRaws:       make([]float64, len(members)),
					Complexities:   make([]float64, len(members)),
					LossInsts:      make([]float64, len(members)),
					LossFinals:     make([]float64, len(members)),
				}
				for i, ind := range members {
					entry.LossRaws[i] = ind.LossRaw
					entry.Complexities[i] = ind.Complexity
					entry.LossInsts[i] = ind.LossInst
					entry.LossFinals[i] = ind.LossFinal
				}
				speciesEntries = append(speciesEntries, entry)
			}
			history.Species[generation] = speciesEntries

			// measure complexity distribution
			complexity_bins := [10]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
			for _, ind := range pop {
				bin := int((ind.Complexity * 0.8) * 10)
				if bin >= 10 {
					bin = 9
				}
				complexity_bins[bin]++
			}
			history.Complexity[generation] = complexity_bins

			// Print species information if verbose
			if verbose >= 2 {
				num := verbose - 1
				tabw := table.NewWriter()
				off_counts_slice := make([]int, 0, min(3, len(offspring_counts)))
				for _, speciesID := range fitness_sorted_ids {
					off_counts_slice = append(off_counts_slice, offspring_counts[speciesID])
				}
				top3 := fmt.Sprintf("%v", off_counts_slice)
				tabw.SetTitle("Species for Generation %d, total Population size (%d=%d+%d%s)", generation, len(pop), elites_count, total_offspring, top3)
				tabw.AppendHeader(table.Row{"ID", "Size", "Off", "Inv", "Fit", "Expression", "LossRaw", "Complexity", "LossInst", "LossFinal"})
				for _, speciesID := range fitness_sorted_ids {
					members := species[speciesID]
					exprs := ""
					lrs := ""
					cs := ""
					lis := ""
					lfns := ""

					add_to_str := func(ind *Individual) {
						tree, err := ind.Tree.ToTree()
						if err != nil {
							fmt.Printf("Error converting tree to string: %v\n", err)
							return
						}
						tressstr := tree.String()
						if len(tressstr) > 50 {
							tressstr = tressstr[:50] + "..."
						}
						exprs += tressstr
						lrs += fmt.Sprintf("%0.3e", ind.LossRaw)
						cs += fmt.Sprintf("%0.3e", ind.Complexity)
						lis += fmt.Sprintf("%0.3e", ind.LossInst)
						lfns += fmt.Sprintf("%0.3e", ind.LossFinal)
					}
					add_to_str(species_representatives[speciesID])
					for i := 0; i+1 < num && i < len(members); i++ {
						exprs += "\n"
						cs += "\n"
						lrs += "\n"
						lis += "\n"
						lfns += "\n"
						add_to_str(members[i])
					}
					tabw.AppendRow(table.Row{speciesID, len(members), offspring_counts[speciesID], species_invalid_counts[speciesID], fmt.Sprintf("%0.3e", species_fitness[speciesID]), exprs, lrs, cs, lis, lfns})
				}
				fmt.Println(tabw.Render() + "\n")
			}
		}

		// Check if we have a new best overall solution
		currentBest := pop[0]
		if history.BestOverall == nil || currentBest.LossFinal < history.BestOverall.LossFinal {
			history.BestOverall = currentBest.Copy()
			history.HallOfFame[generation] = currentBest.Copy()

			if verbose >= 1 {
				tree, err := currentBest.Tree.ToTree()
				if err != nil {
					fmt.Printf("Error converting tree to string: %v\n", err)
					continue
				}
				tressstr := tree.String()
				if len(tressstr) > 50 {
					tressstr = tressstr[:50] + "..."
				}
				log := fmt.Sprintf("Gen %d, Species %d, Loss %0.3e, %s", generation, len(species), currentBest.LossFinal, tressstr)
				if verbose >= 2 {
					fmt.Println(log)
				} else {
					pbar.Message = log
				}
				if len(currentBest.Tree) > 30 {
					fmt.Printf("Note: Expression is quite long (%d nodes).\n", len(currentBest.Tree))
					fmt.Printf("LossRaw = %0.3f, Complexity = %0.3f, LossInst = %0.3f, LossFinal = %0.3f\n\n", history.BestOverall.LossRaw, history.BestOverall.Complexity, history.BestOverall.LossInst, history.BestOverall.LossFinal)
				}
			}
		}

		// Create new population
		newPop := make(Population, 0, conf.PopulationSize)

		// Elitism Carry-over
		top_elite_count := min(conf.TopElites, elites_count)
		for i := 0; i < top_elite_count && i < len(pop); i++ {
			//t, _ := pop[i].Tree.ToTree()
			//fmt.Printf("Carrying over top elite individual %d with Loss = %g, Expression = %s\n", i+1, pop[i].LossFinal, t.String())
			newPop = append(newPop, pop[i].Copy())
		}
		i := 0
		for len(newPop) < elites_count {
			speciesID := fitness_sorted_ids[i%len(fitness_sorted_ids)]
			start_idx := min(i/len(fitness_sorted_ids)*conf.SpeciesElites, len(species[speciesID]))
			end_idx := min(start_idx+conf.SpeciesElites, len(species[speciesID]))
			for _, ind := range species[speciesID][start_idx:end_idx] {
				//t, _ := ind.Tree.ToTree()
				//fmt.Printf("Carrying over elite from species %d with Loss = %g, Expression = %s\n", speciesID, ind.LossFinal, t.String())
				newPop = append(newPop, ind.Copy())
				if len(newPop) >= elites_count {
					break
				}
			}
			i++
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
		for speciesID, members := range species {
			added := 0
			// For each species, add offspring until we reach the allocated number of offspring for this species
			for added < offspring_counts[speciesID] {
				mother := select_individual(members)
				if rnd.Float64() < conf.CrossoverRate {
					// Crossover
					var father *Individual
					if rnd.Float64() < conf.InterSpeciesMatingRate {
						// select random father from any species
						father = select_individual(pop)
					} else {
						// select father from the same species
						father = select_individual(members)
					}
					off1, off2 := Crossover(mother, father)
					maybe_mutate_and_add(off1)
					added++
					// check if limit is reached
					if added < offspring_counts[speciesID] {
						maybe_mutate_and_add(off2)
						added++
					}
				} else {
					// Pure mutation
					maybe_mutate_and_add(mother)
					added++
				}
			}
		}
		pop = newPop
		pbar.Increment(1)
	}
	history.GenerationsTakenForOptimal = -1
	for gen := range history.HallOfFame {
		if gen > history.GenerationsTakenForOptimal {
			history.GenerationsTakenForOptimal = gen
		}
	}
	bestOverallTree, err := history.BestOverall.Tree.ToTree()
	if err != nil {
		fmt.Printf("Error converting best overall tree to string: %v\n", err)
	}
	history.BestOverallTree = bestOverallTree
	if verbose >= 1 {
		fmt.Printf("Best solution found in generation %d with Loss = %g, Expression = %s\n", history.GenerationsTakenForOptimal, history.BestOverall.LossFinal, bestOverallTree.String())
	}
	return
}
