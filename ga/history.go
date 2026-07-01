package ga

import (
	"encoding/json"
	"marvin/symbreggenalgo/symbolic"
	"os"
)

type EvolutionHistory struct {
	GenerationsTakenForOptimal int
	BestOverall                *Individual
	BestOverallTree            *symbolic.Tree
	HallOfFame                 map[int]*Individual
	Complexity                 map[int][10]int
	Species                    map[int][]SpeciesEntry
}

type SpeciesEntry struct {
	ID             int
	Size           int
	Representative symbolic.Postfix
	LossRaws       []float64
	Complexities   []float64
	LossInsts      []float64
	LossFinals     []float64
}

// Serializable versions of your structs
type SpeciesEntryDTO struct {
	ID             int       `json:"id"`
	Size           int       `json:"size"`
	Representative string    `json:"representative"`
	LossRaws       []float64 `json:"loss_raws"`
	Complexities   []float64 `json:"complexities"`
	LossInsts      []float64 `json:"loss_insts"`
	LossFinals     []float64 `json:"loss_finals"`
}

type EvolutionHistoryDTO struct {
	GenerationsTakenForOptimal int                       `json:"generations_taken_for_optimal"`
	BestOverall                string                    `json:"best_overall"`
	HallOfFame                 map[int]string            `json:"hall_of_fame"`
	Complexity                 map[int][10]int           `json:"complexity"`
	Species                    map[int][]SpeciesEntryDTO `json:"species"`
}

func ExportHistoryToJSON(history *EvolutionHistory, filename string) error {
	dto := EvolutionHistoryDTO{
		GenerationsTakenForOptimal: history.GenerationsTakenForOptimal,
		HallOfFame:                 make(map[int]string),
		Complexity:                 history.Complexity,
		Species:                    make(map[int][]SpeciesEntryDTO),
	}

	// Safely convert BestOverall to string
	if history.BestOverall != nil {
		if tree, err := history.BestOverall.Tree.ToTree(); err == nil {
			dto.BestOverall = tree.String()
		}
	}

	// Convert HallOfFame
	for gen, ind := range history.HallOfFame {
		if tree, err := ind.Tree.ToTree(); err == nil {
			dto.HallOfFame[gen] = tree.Stringfmt("%v")
		}
	}

	// Convert Species
	for gen, entries := range history.Species {
		var dtoEntries []SpeciesEntryDTO
		for _, entry := range entries {
			// Convert Postfix (Representative) to string
			tree, _ := entry.Representative.ToTree()

			dtoEntries = append(dtoEntries, SpeciesEntryDTO{
				ID:             entry.ID,
				Size:           entry.Size,
				Representative: tree.Stringfmt("%v"), // Or entry.Representative.String() if implemented
				LossRaws:       entry.LossRaws,
				Complexities:   entry.Complexities,
				LossInsts:      entry.LossInsts,
				LossFinals:     entry.LossFinals,
			})
		}
		dto.Species[gen] = dtoEntries
	}

	// Serialize to JSON
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Nice formatting
	return encoder.Encode(dto)
}
