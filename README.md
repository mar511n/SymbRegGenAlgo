# Symbolic Regression Genetic Algorithm

This project implements a genetic algorithm in Go to perform symbolic regression. It evolves mathematical expressions to fit datasets by mimicking biological evolution processes.

## Features

- **Expression Representation:** Uses binary trees to represent and evolve mathematical functions.
- **Speciation:** Maintains population diversity by grouping similar individuals, preventing premature convergence.
- **Elitism:** Ensures the best-performing individuals are carried over to the next generation.
- **Genetic Operators:** Includes mutation and crossover to explore the solution space.
- **Dataset Management:** Logic to generate synthetic datasets or load existing ones.
- **History Export:** Evolution progress and statistics are exported as JSON for post-run analysis.

## Project Structure

- `main.go`: Entry point for dataset generation, evolution execution, and JSON export.
- `analysis.ipynb`: A Jupyter Notebook designed to read the JSON history files and visualize the results.

## Usage

1. **Run Evolution**:
	```bash
	go run main.go datagen.go
	```
2. **Analyze Results**:
	Open `analysis.ipynb` in a Jupyter environment to load the generated JSON data and inspect fitness curves and evolved expressions.