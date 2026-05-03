package main

import (
	"encoding/csv"
	"fmt"
	"marvin/symbreggenalgo/ga"
	"marvin/symbreggenalgo/symbolic"
	"math"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/schollz/progressbar/v3"
)

func GetProgressBar(N int) *progressbar.ProgressBar {
	return progressbar.NewOptions64(
		int64(N),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(10),
		progressbar.OptionShowTotalBytes(true),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() {
			// Clear the progress bar line on completion
			fmt.Fprint(os.Stderr, "\r\x1b[2K") // Clear line and return carriage
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)
}

func min(slice []float64) float64 {
	min := math.Inf(1)
	for _, v := range slice {
		if v < min {
			min = v
		}
	}
	return min
}

func max(slice []float64) float64 {
	max := math.Inf(-1)
	for _, v := range slice {
		if v > max {
			max = v
		}
	}
	return max
}

func mean(slice []float64) float64 {
	sum := 0.0
	for _, v := range slice {
		sum += v
	}
	return sum / float64(len(slice))
}

func median(slice []float64) float64 {
	sorted := make([]float64, len(slice))
	copy(sorted, slice)
	sort.Float64s(sorted)
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func main() {
	//generateAllDatasets()
	datasetFile := "data/datasets/dataset_depth2_vars2_2.csv"
	conf := ga.DefaultConfig()
	conf.PopulationSize = 100
	conf.Generations = 1000

	conf.CompatibilityThreshold = 0.41

	conf.MaxLossRaw = -1 //0.05 //-1 means, the max loss is guessed from the initial population
	conf.MaxComplexity = 7
	conf.MinComplexityWeight = 0.05
	conf.MaxComplexityWeight = 0.2

	conf.UsedSelection = ga.Tournament
	conf.SelectionParams = 3
	conf.InterSpeciesMatingRate = 0.8
	conf.MutationRate = 0.9
	conf.CrossoverRate = 0.7
	conf.GlobalElitismCount = 4
	conf.SpeciesElites = 1
	conf.TopElites = 1

	_, history := runGeneticAlgorithm(datasetFile, conf, 2, 100)

	// Save history to JSON file
	ga.ExportHistoryToJSON(history, "data/results/evolution_history.json")
}

/* func parameter_sweep_penalty_size() {
	// Run the GA on a specific dataset
	datasetFile := "datasets/dataset_depth2_vars1_4.csv"
	// Number of runs per penalty size
	N := 20

	// Target expression for reference
	testcfg := ga.DefaultConfig()
	testcfg.Generations = 1
	testcfg.PopulationSize = 1
	_, _, _, _, target_tree, _ := runGeneticAlgorithm(datasetFile, testcfg, false, 0)
	fmt.Printf("Target expression: %s\n", target_tree.String())

	// Scanned penalty sizes
	min_ps := 0.1
	max_ps := 4.0
	num_ps := 10
	penalty_size := make([]float64, 0, num_ps)
	for i := range num_ps {
		penalty_size = append(penalty_size, min_ps+float64(i)*(max_ps-min_ps)/float64(num_ps-1))
	}

	score_distributions := make([][]float64, len(penalty_size))
	expression_size_distributions := make([][]int32, len(penalty_size))
	for i, ps := range penalty_size {
		scores := make([]float64, 0, N)
		expression_sizes := make([]int32, 0, N)
		bar := GetProgressBar(N)
		bestLoss := math.Inf(1)
		bestTree := &symbolic.Tree{}

		for range N {
			conf := ga.DefaultConfig()
			conf.PenaltySize = ps
			score, _, individual, tree, _, _ := runGeneticAlgorithm(datasetFile, conf, false, 0)
			if individual.Loss < bestLoss {
				bestLoss = individual.Loss
				bestTree = tree
			}
			scores = append(scores, score)
			expression_sizes = append(expression_sizes, int32(len(individual.Tree)))
			bar.Add(1)
		}

		fmt.Printf("(%v/%v) Penalty size: %0.2f, Score: min=%0.2f, max=%0.2f, mean=%0.2f, median=%0.2f, Fit: %0.2f, Expr: %s\n", i+1, len(penalty_size), ps, min(scores), max(scores), mean(scores), median(scores), bestLoss, bestTree.String())

		score_distributions[i] = scores
		expression_size_distributions[i] = expression_sizes
	}

	writer, err := gonpy.NewFileWriter("penalty_sizes.npy")
	if err != nil {
		fmt.Println("Error creating numpy file:", err)
		return
	}
	writer.Shape = []int{len(penalty_size)}
	err = writer.WriteFloat64(penalty_size)
	if err != nil {
		fmt.Println("Error writing to numpy file:", err)
		return
	}
	fmt.Printf("Saved penalty sizes to penalty_sizes.npy\n")

	writer, err = gonpy.NewFileWriter("penalty_size_scores.npy")
	if err != nil {
		fmt.Println("Error creating numpy file:", err)
		return
	}
	writer.Shape = []int{len(penalty_size), N}
	data := make([]float64, 0, len(penalty_size)*N)
	for i := range penalty_size {
		data = append(data, score_distributions[i]...)
	}
	err = writer.WriteFloat64(data)
	if err != nil {
		fmt.Println("Error writing to numpy file:", err)
		return
	}
	fmt.Printf("Saved score distributions to penalty_size_scores.npy\n")

	writer, err = gonpy.NewFileWriter("penalty_size_expression_sizes.npy")
	if err != nil {
		fmt.Println("Error creating numpy file:", err)
		return
	}
	writer.Shape = []int{len(penalty_size), N}
	exprData := make([]int32, 0, len(penalty_size)*N)
	for i := range penalty_size {
		exprData = append(exprData, expression_size_distributions[i]...)
	}
	err = writer.WriteInt32(exprData)
	if err != nil {
		fmt.Println("Error writing to numpy file:", err)
		return
	}
	fmt.Printf("Saved expression size distributions to penalty_size_expression_sizes.npy\n")
}
*/

func runGeneticAlgorithm(datasetFile string, conf *ga.Config, run_verbose int, num_hist_dumps int) (score float64, history *ga.EvolutionHistory) {
	f, err := os.Open(datasetFile)
	if err != nil {
		fmt.Println("Error opening dataset:", err)
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading dataset:", err)
		return
	}

	if len(records) < 2 {
		fmt.Println("Dataset is empty or has no data rows")
		return
	}

	header := records[0]
	varNames := header[1:]
	targetPostfix, err := symbolic.ParsePostfix(header[0])
	if err != nil {
		fmt.Printf("Error parsing header expression: %v\n", err)
		return
	}
	targetTree, err := targetPostfix.ToTree()
	if err != nil {
		fmt.Printf("Error converting header postfix to tree: %v\n", err)
		return
	}

	var dataset ga.Dataset
	for i := 1; i < len(records); i++ {
		row := records[i]
		target, _ := strconv.ParseFloat(row[0], 64)
		vars := make(map[string]float64)
		for j, varName := range varNames {
			val, _ := strconv.ParseFloat(row[j+1], 64)
			vars[varName] = val
		}
		dataset = append(dataset, ga.DataPoint{
			Target:    target,
			Variables: vars,
		})
	}

	alphabet := ga.DefaultAlphabet(varNames)
	history = ga.Run(dataset, conf, alphabet, run_verbose, num_hist_dumps)

	targetIndividual := &ga.Individual{Tree: targetPostfix}
	ga.EvaluateLoss(targetIndividual, dataset, conf, 1.0, 1)
	score = targetIndividual.LossFinal / history.BestOverall.LossFinal

	tabw := table.NewWriter()
	tabw.SetTitle("Evolution Results")
	tabw.AppendHeader(table.Row{"Type", "LossRaw", "Complexity", "LossInst", "LossFinal", "Expression"})
	tabw.AppendRow(table.Row{"Target", fmt.Sprintf("%0.3e", targetIndividual.LossRaw), fmt.Sprintf("%0.3e", targetIndividual.Complexity), fmt.Sprintf("%0.3e", targetIndividual.LossInst), fmt.Sprintf("%0.3e", targetIndividual.LossFinal), targetTree.String()})
	tabw.AppendRow(table.Row{"Best", fmt.Sprintf("%0.3e", history.BestOverall.LossRaw), fmt.Sprintf("%0.3e", history.BestOverall.Complexity), fmt.Sprintf("%0.3e", history.BestOverall.LossInst), fmt.Sprintf("%0.3e", history.BestOverall.LossFinal), history.BestOverallTree.String()})
	tabw.SetCaption("Score (Target Loss / Best Loss) = %0.4f", score)
	fmt.Println(tabw.Render())
	return
}
