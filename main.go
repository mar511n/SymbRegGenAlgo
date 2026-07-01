package main

import (
	"encoding/csv"
	"encoding/json"
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

// Config represents the parameters loaded from a JSON config file.
type Config struct {
	DatasetFilepath string       `json:"dataset_filepath"`
	HistoryFilepath string       `json:"history_filepath"`
	PlotFilepath    string       `json:"plot_filepath"`
	Vars            []string     `json:"vars"`
	VarRanges       [][2]float64 `json:"var_ranges"`
	ExprStr         string       `json:"expr_str"`
	NumDatapoints   int          `json:"num_datapoints"`
	NoiseStddev     float64      `json:"noise_stddev"`
	NumHistoryDumps int          `json:"num_history_dumps"`
}

func loadConfig(filepath string) (*Config, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GAConfig is a JSON-serializable representation of ga.Config.
type GAConfig struct {
	PopulationSize         int                   `json:"population_size"`
	Generations            int                   `json:"generations"`
	MaxDepth               int                   `json:"max_depth"`
	CrossoverRate          float64               `json:"crossover_rate"`
	MutationRate           float64               `json:"mutation_rate"`
	MaxLossRaw             float64               `json:"max_loss_raw"`
	MaxComplexity          float64               `json:"max_complexity"`
	MinComplexityWeight    float64               `json:"min_complexity_weight"`
	MaxComplexityWeight    float64               `json:"max_complexity_weight"`
	UsedSelection          string                `json:"used_selection"`
	SelectionParams        int                   `json:"selection_params"`
	GlobalElitismCount     int                   `json:"global_elitism_count"`
	SpeciesElites          int                   `json:"species_elites"`
	TopElites              int                   `json:"top_elites"`
	CompatibilityThreshold float64               `json:"compatibility_threshold"`
	InterSpeciesMatingRate float64               `json:"inter_species_mating_rate"`
	DifferenceMeasure      DifferenceMeasureJSON `json:"difference_measure"`
	GeneratorParams        *GeneratorParamsJSON  `json:"generator_params"`
}

type DifferenceMeasureJSON struct {
	TreeSizeWeight  float64 `json:"tree_size_weight"`
	TokenDiffWeight float64 `json:"token_diff_weight"`
}

type GeneratorParamsJSON struct {
	EarlyTerminationProb float64            `json:"early_termination_prob"`
	BinaryProb           float64            `json:"binary_prob"`
	ConstantProb         float64            `json:"constant_prob"`
	BinaryOpWeights      map[string]float64 `json:"binary_op_weights"`
	UnaryOpWeights       map[string]float64 `json:"unary_op_weights"`
	PointMutationProb    float64            `json:"point_mutation_prob"`
	LeafGrowthProb       float64            `json:"leaf_growth_prob"`
	RootGrowthProb       float64            `json:"root_growth_prob"`
}

var binaryOpNames = map[string]symbolic.BinaryOp{
	"Add": symbolic.Add,
	"Sub": symbolic.Sub,
	"Mul": symbolic.Mul,
	"Div": symbolic.Div,
	"Pow": symbolic.Pow,
	"Max": symbolic.Max,
	"Min": symbolic.Min,
	"Mod": symbolic.Mod,
}

var unaryOpNames = map[string]symbolic.UnaryOp{
	"Sin":   symbolic.Sin,
	"Cos":   symbolic.Cos,
	"Exp":   symbolic.Exp,
	"Log":   symbolic.Log,
	"Abs":   symbolic.Abs,
	"Floor": symbolic.Floor,
	"Asin":  symbolic.Asin,
	"Acos":  symbolic.Acos,
	"Atan":  symbolic.Atan,
}

func (g *GAConfig) ToGaConfig() *ga.Config {
	conf := ga.DefaultConfig()
	conf.PopulationSize = g.PopulationSize
	conf.Generations = g.Generations
	conf.MaxDepth = g.MaxDepth
	conf.CrossoverRate = g.CrossoverRate
	conf.MutationRate = g.MutationRate
	conf.MaxLossRaw = g.MaxLossRaw
	conf.MaxComplexity = g.MaxComplexity
	conf.MinComplexityWeight = g.MinComplexityWeight
	conf.MaxComplexityWeight = g.MaxComplexityWeight
	switch g.UsedSelection {
	case "WeightedLoss":
		conf.UsedSelection = ga.WeightedLoss
	default:
		conf.UsedSelection = ga.Tournament
	}
	conf.SelectionParams = g.SelectionParams
	conf.GlobalElitismCount = g.GlobalElitismCount
	conf.SpeciesElites = g.SpeciesElites
	conf.TopElites = g.TopElites
	conf.CompatibilityThreshold = g.CompatibilityThreshold
	conf.InterSpeciesMatingRate = g.InterSpeciesMatingRate

	// DifferenceMeasure
	conf.DifferenceMeasure.TreeSizeWeight = g.DifferenceMeasure.TreeSizeWeight
	conf.DifferenceMeasure.TokenDiffWeight = g.DifferenceMeasure.TokenDiffWeight

	// GeneratorParams
	if g.GeneratorParams != nil {
		gen := ga.DefaultGeneratorParams()
		gen.EarlyTerminationProb = g.GeneratorParams.EarlyTerminationProb
		gen.BinaryProb = g.GeneratorParams.BinaryProb
		gen.ConstantProb = g.GeneratorParams.ConstantProb
		gen.PointMutationProb = g.GeneratorParams.PointMutationProb
		gen.LeafGrowthProb = g.GeneratorParams.LeafGrowthProb
		gen.RootGrowthProb = g.GeneratorParams.RootGrowthProb

		// Convert string-keyed weights back to symbolic op maps
		if g.GeneratorParams.BinaryOpWeights != nil {
			gen.BinaryOpWeights = make(map[symbolic.BinaryOp]float64)
			for name, w := range g.GeneratorParams.BinaryOpWeights {
				if op, ok := binaryOpNames[name]; ok {
					gen.BinaryOpWeights[op] = w
				}
			}
		}
		if g.GeneratorParams.UnaryOpWeights != nil {
			gen.UnaryOpWeights = make(map[symbolic.UnaryOp]float64)
			for name, w := range g.GeneratorParams.UnaryOpWeights {
				if op, ok := unaryOpNames[name]; ok {
					gen.UnaryOpWeights[op] = w
				}
			}
		}

		conf.Params = gen
	}

	return conf
}

func loadGaConfig(filepath string) (*GAConfig, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg GAConfig
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go datagen.go <config_file> <ga_config_file>")
		fmt.Println("Example: go run main.go datagen.go config.json ga_config.json")
		os.Exit(1)
	}

	cfg, err := loadConfig(os.Args[1])
	if err != nil {
		fmt.Printf("Error loading config file %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	gaCfg, err := loadGaConfig(os.Args[2])
	if err != nil {
		fmt.Printf("Error loading GA config file %s: %v\n", os.Args[2], err)
		os.Exit(1)
	}

	dataset_filepath := cfg.DatasetFilepath
	history_filepath := cfg.HistoryFilepath
	plot_filepath := cfg.PlotFilepath
	vars := cfg.Vars
	var_ranges := cfg.VarRanges
	expr_str := cfg.ExprStr
	num_datapoints := cfg.NumDatapoints
	noise_stddev := cfg.NoiseStddev
	num_history_dumps := cfg.NumHistoryDumps

	postfix, err := symbolic.ParsePostfix(expr_str)
	if err != nil {
		fmt.Println("Error parsing postfix expression:", err)
		return
	}
	generateNoisyDataset(dataset_filepath, vars, var_ranges, postfix, num_datapoints, noise_stddev)

	conf := gaCfg.ToGaConfig()

	_, target, data, history := runGeneticAlgorithm(dataset_filepath, conf, 2, num_history_dumps)

	// Save history to JSON file
	ga.ExportHistoryToJSON(history, history_filepath)

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

		err := p.Save(10*vg.Inch, 8*vg.Inch, plot_filepath)
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
