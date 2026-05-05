package main

import (
	"encoding/csv"
	"fmt"
	"image/color"
	"marvin/symbreggenalgo/ga"
	"marvin/symbreggenalgo/symbolic"
	"math"
	"os"
	"sort"
	"strconv"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

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
	conf.PopulationSize = 200
	conf.Generations = 300

	conf.CompatibilityThreshold = 0.505

	conf.MaxLossRaw = -1 //0.05 //-1 means, the max loss is guessed from the initial population
	conf.MaxComplexity = 10
	conf.MinComplexityWeight = 0.05
	conf.MaxComplexityWeight = 0.05

	conf.UsedSelection = ga.Tournament
	conf.SelectionParams = 2
	conf.InterSpeciesMatingRate = 0.7
	conf.MutationRate = 0.7
	conf.CrossoverRate = 0.5
	conf.GlobalElitismCount = 11
	conf.SpeciesElites = 2
	conf.TopElites = 1

	conf.MaxDepth = 1

	// TODO: there is a bug, where equal!? (0.00e+00 or NaN) individuals are not grouped into the same species. Check distance measure and species assignment.
	_, target, data, history := runGeneticAlgorithm(datasetFile, conf, 2, 100)

	// Save history to JSON file
	ga.ExportHistoryToJSON(history, "data/results/evolution_history.json")

	// create plot of best expression together with data and target expression
	history.BestOverall.Evaluate(data)
	best_predictions := history.BestOverall.GetPredictions()
	target.Evaluate(data)
	target_predictions := target.GetPredictions()
	b_vars := history.BestOverall.Tree.ContainedVariables()
	t_vars := target.Tree.ContainedVariables()
	if len(b_vars) > 1 || len(t_vars) > 1 {
		fmt.Printf("Can only plot expressions with one variable...\n")
	} else {
		x := make([]float64, len(data))
		y := make([]float64, len(data))
		for i, dp := range data {
			x[i] = dp.Variables[b_vars[0]]
			y[i] = dp.Target
		}

		type SortPoint struct {
			x, y, best, target float64
		}
		spts := make([]SortPoint, len(data))
		for i := range x {
			spts[i] = SortPoint{x: x[i], y: y[i], best: best_predictions[i], target: target_predictions[i]}
		}
		sort.Slice(spts, func(i, j int) bool { return spts[i].x < spts[j].x })

		pts := make(plotter.XYs, len(x))
		bestPts := make(plotter.XYs, len(x))
		targetPts := make(plotter.XYs, len(x))
		for i, sp := range spts {
			pts[i].X, pts[i].Y = sp.x, sp.y
			bestPts[i].X, bestPts[i].Y = sp.x, sp.best
			targetPts[i].X, targetPts[i].Y = sp.x, sp.target
		}

		p := plot.New()
		p.Title.Text = "Symbolic Regression Results"
		p.X.Label.Text = "x"
		p.Y.Label.Text = "y"

		scatter, _ := plotter.NewScatter(pts)
		p.Add(scatter)

		bestLine, _ := plotter.NewLine(bestPts)
		bestLine.LineStyle.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
		p.Add(bestLine)

		targetLine, _ := plotter.NewLine(targetPts)
		targetLine.LineStyle.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}
		p.Add(targetLine)

		p.Legend.Add("Data", scatter)
		p.Legend.Add("Best", bestLine)
		p.Legend.Add("Target", targetLine)

		err := p.Save(10*vg.Inch, 8*vg.Inch, "data/results/predictions.svg")
		if err != nil {
			fmt.Println("Error saving plot:", err)
		}
	}

}

func runGeneticAlgorithm(datasetFile string, conf *ga.Config, run_verbose int, num_hist_dumps int) (score float64, targetIndividual *ga.Individual, dataset ga.Dataset, history *ga.EvolutionHistory) {
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

	targetIndividual = &ga.Individual{Tree: targetPostfix}
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
