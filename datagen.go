package main

import (
	"encoding/csv"
	"fmt"
	"marvin/symbreggenalgo/ga"
	"marvin/symbreggenalgo/symbolic"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
)

func setVariables(n symbolic.Node, vars map[string]float64) {
	if n == nil {
		return
	}
	switch node := n.(type) {
	case *symbolic.InputNode:
		if val, ok := vars[node.Name]; ok {
			node.Value = val
		}
	case *symbolic.UnaryNode:
		setVariables(node.Input, vars)
	case *symbolic.BinaryNode:
		setVariables(node.Left, vars)
		setVariables(node.Right, vars)
	}
}

func addNoise(val float64, stdDev float64) float64 {
	return val * (1 + rand.NormFloat64()*stdDev)
}

// GenerateRandomDatasets generates multiple random expressions and creates a dataset for each.
func GenerateRandomDatasets(numDatasets int, numPoints int, maxDepth int, numVariables int, outDir, filename string, genparams *ga.GeneratorParams, minConst, maxConst float64, noiseStdDev float64) error {
	vars := make([]string, numVariables)
	for i := 0; i < numVariables; i++ {
		vars[i] = fmt.Sprintf("x%d", i)
	}
	alphabet := ga.DefaultAlphabet(vars)
	alphabet.MinConst = minConst
	alphabet.MaxConst = maxConst

	err := os.MkdirAll(outDir, os.ModePerm)
	if err != nil {
		return err
	}

	generated := 0
	for generated < numDatasets {
		postfix := ga.GenerateTree(maxDepth, alphabet, genparams)
		tree, err := postfix.ToTree()
		if err != nil {
			continue // try next if invalid tree
		}

		postfixStr := postfix.String()
		filename := filepath.Join(outDir, fmt.Sprintf("%s_%d.csv", filename, generated))
		f, err := os.Create(filename)
		if err != nil {
			return err
		}

		writer := csv.NewWriter(f)

		// Write header: first column is the expression string (resulting output), then the variables
		header := []string{postfixStr}
		header = append(header, vars...)
		writer.Write(header)

		validPoints := 0
		attempts := 0
		maxAttempts := numPoints * 10

		for validPoints < numPoints && attempts < maxAttempts {
			attempts++

			inputs := make(map[string]float64)
			noisyInputs := make(map[string]float64)
			for _, v := range vars {
				val := (rand.Float64() * 20) - 10 // [-10, 10] range
				inputs[v] = val
				noisyInputs[v] = addNoise(val, noiseStdDev)
			}

			// Evaluate using exact parameters
			setVariables(tree.Root, inputs)
			out := tree.Evaluate()

			// If result is NaN or Inf, skip
			if math.IsNaN(out) || math.IsInf(out, 0) {
				continue
			}

			// Add measurement error to output
			noisyOut := addNoise(out, noiseStdDev)

			// Write row
			row := []string{strconv.FormatFloat(noisyOut, 'f', -1, 64)}
			for _, v := range vars {
				row = append(row, strconv.FormatFloat(noisyInputs[v], 'f', -1, 64))
			}
			writer.Write(row)
			validPoints++
		}

		writer.Flush()
		f.Close()

		// If we couldn't generate enough valid points, delete the file and don't count it
		if validPoints < numPoints {
			os.Remove(filename)
			continue
		}

		generated++
	}
	return nil
}

func generateAllDatasets() {
	fmt.Println("Generating datasets...")
	max_depths := []int{2, 3}
	num_variables := []int{1, 2}
	for _, depth := range max_depths {
		for _, vars := range num_variables {
			genparams := ga.DefaultGeneratorParams()
			err := GenerateRandomDatasets(
				20,              // number of datasets per configuration
				100,             // number of data points per dataset
				depth,           // maximum depth of the generated expression tree
				vars,            // number of variables
				"data/datasets", // output directory
				fmt.Sprintf("dataset_depth%d_vars%d", depth, vars), // filename prefix
				genparams, // GeneratorParams with default settings
				-10.0,     // min constant value
				10.0,      // max constant value
				0.01,      // 10% noise
			)
			if err != nil {
				fmt.Println("Error:", err)
			}
		}
	}
	fmt.Println("Dataset generation complete.")
}
