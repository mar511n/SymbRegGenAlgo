package ga

import (
	"marvin/symbreggenalgo/symbolic"
	"math"
)

// FindSubtreeBounds returns the start index of the subtree resolving at endIndex.
func FindSubtreeBounds(p symbolic.Postfix, endIndex int) int {
	needed := 1
	for i := endIndex; i >= 0; i-- {
		switch p[i].Type {
		case symbolic.TokenTypeBinary:
			needed += 1 // requires 2 inputs, acts as 1 output (net +1 backwards)
		case symbolic.TokenTypeUnary:
			// requires 1, outputs 1 (net 0)
		default: // Constant / Variable
			needed -= 1 // provides an input
		}
		if needed == 0 {
			return i
		}
	}
	return 0 // Fallback
}

// Crossover swaps random subtrees between two parents
func Crossover(parent1, parent2 *Individual) (*Individual, *Individual) {
	p1 := parent1.Tree
	p2 := parent2.Tree

	if len(p1) == 0 || len(p2) == 0 {
		return parent1.Copy(), parent2.Copy()
	}

	cut1 := rnd.Intn(len(p1))
	cut2 := rnd.Intn(len(p2))

	start1 := FindSubtreeBounds(p1, cut1)
	start2 := FindSubtreeBounds(p2, cut2)

	off1 := append(symbolic.Postfix{}, p1[:start1]...)
	off1 = append(off1, p2[start2:cut2+1]...)
	off1 = append(off1, p1[cut1+1:]...)

	off2 := append(symbolic.Postfix{}, p2[:start2]...)
	off2 = append(off2, p1[start1:cut1+1]...)
	off2 = append(off2, p2[cut2+1:]...)

	return &Individual{Tree: off1.Simplify()}, &Individual{Tree: off2.Simplify()}
}

// Mutate implements Point Mutation, Subtree Replacement, OR Shrink Mutation randomly.
func Mutate(ind *Individual, alpha *Alphabet, maxDepth int, params *GeneratorParams) {
	if len(ind.Tree) == 0 {
		return
	}

	randVal := rnd.Float64()
	if randVal < params.PointMutationProb {
		// Point Mutation (randomly change an operator, variable, or constant)
		PointMutation(ind, alpha)
	} else if randVal < params.PointMutationProb+params.LeafGrowthProb {
		// Leaf Growth Mutation (randomly replace a subtree with a new randomly generated subtree)
		GrowLeafMutation(ind, alpha, maxDepth, params)
	} else if randVal < params.PointMutationProb+params.LeafGrowthProb+params.RootGrowthProb {
		// Root Growth Mutation (generate a new random tree and insert the old tree as a subtree at a random point)
		GrowRootMutation(ind, alpha, maxDepth, params)
	} else {
		// Shrink Mutation (Either removes a random unary operator by replacing it with its child or replaces a random subtree with a terminal)
		if rnd.Float64() < 0.5 {
			RemoveUnaryMutation(ind)
		} else {
			ShrinkMutation(ind, alpha, maxDepth, params)
		}
	}
	ind.Tree = ind.Tree.Simplify()
}

func PointMutation(ind *Individual, alpha *Alphabet) {
	idx := rnd.Intn(len(ind.Tree))
	token := &ind.Tree[idx]
	switch token.Type {
	case symbolic.TokenTypeBinary:
		// Randomly select a different binary operator
		if len(alpha.BinaryOps) > 0 {
			token.BinaryOp = alpha.BinaryOps[rnd.Intn(len(alpha.BinaryOps))]
		}
	case symbolic.TokenTypeUnary:
		// Randomly select a different unary operator
		if len(alpha.UnaryOps) > 0 {
			token.UnaryOp = alpha.UnaryOps[rnd.Intn(len(alpha.UnaryOps))]
		}
	case symbolic.TokenTypeConstant:
		// Shift the constant value by random amount
		if rnd.Float64() < 0.5 {
			token.Value = alpha.MinConst + rnd.NormFloat64()*(alpha.MaxConst-alpha.MinConst)/2/3 + (alpha.MaxConst+alpha.MinConst)/2
		} else {
			token.Value = alpha.MinConst + rnd.Float64()*(alpha.MaxConst-alpha.MinConst)
		}
	case symbolic.TokenTypeVariable:
		// Randomly select a different variable
		if len(alpha.Variables) > 0 {
			token.Name = alpha.Variables[rnd.Intn(len(alpha.Variables))]
		}
	}
}

func GrowLeafMutation(ind *Individual, alpha *Alphabet, maxDepth int, params *GeneratorParams) {
	// select random subtree to replace
	cut := rnd.Intn(len(ind.Tree))
	start := FindSubtreeBounds(ind.Tree, cut)

	// generate new subtree to insert
	newSubtree := GenerateTree(maxDepth, alpha, params)

	// Replace the subtree at the cut point with the new subtree
	newTree := append(symbolic.Postfix{}, ind.Tree[:start]...)
	newTree = append(newTree, newSubtree...)
	newTree = append(newTree, ind.Tree[cut+1:]...)

	ind.Tree = newTree
}

func GrowRootMutation(ind *Individual, alpha *Alphabet, maxDepth int, params *GeneratorParams) {
	// generate a new random tree
	newTree := GenerateTree(maxDepth, alpha, params)

	// Select a random subtree of the new tree to replace with the old tree
	cut := rnd.Intn(len(newTree))
	start := FindSubtreeBounds(newTree, cut)

	// Replace the subtree at the cut point with the old tree
	grownTree := append(symbolic.Postfix{}, newTree[:start]...)
	grownTree = append(grownTree, ind.Tree...)
	grownTree = append(grownTree, newTree[cut+1:]...)

	ind.Tree = grownTree
}

func RemoveUnaryMutation(ind *Individual) {
	// Select a random unary operator to remove
	unaryIndices := make([]int, 0)
	for i, token := range ind.Tree {
		if token.Type == symbolic.TokenTypeUnary {
			unaryIndices = append(unaryIndices, i)
		}
	}
	if len(unaryIndices) == 0 {
		return // No unary operators to remove
	}
	cut := unaryIndices[rnd.Intn(len(unaryIndices))]
	// check to left and right of the cut for more unary operators and remove them maybe
	lcut := cut
	for lcut > 0 && ind.Tree[lcut-1].Type == symbolic.TokenTypeUnary {
		if rnd.Float64() < 0.5 {
			lcut -= 1
		} else {
			break
		}
	}
	rcut := cut
	for rcut < len(ind.Tree)-1 && ind.Tree[rcut+1].Type == symbolic.TokenTypeUnary {
		if rnd.Float64() < 0.5 {
			rcut += 1
		} else {
			break
		}
	}
	newTree := append(symbolic.Postfix{}, ind.Tree[:lcut]...)
	newTree = append(newTree, ind.Tree[rcut+1:]...)
	ind.Tree = newTree
}

func ShrinkMutation(ind *Individual, alpha *Alphabet, maxDepth int, params *GeneratorParams) {
	// Select a random subtree to replace with a NaN constant (which will be removed by the simplification step)
	cut := rnd.Intn(len(ind.Tree))
	start := FindSubtreeBounds(ind.Tree, cut)

	newTree := append(symbolic.Postfix{}, ind.Tree[:start]...)
	newTree = append(newTree, symbolic.Token{Type: symbolic.TokenTypeConstant, Value: math.NaN()})
	newTree = append(newTree, ind.Tree[cut+1:]...)

	ind.Tree = newTree
	/*
		// Select a random subtree to replace with a terminal
		cut := rnd.Intn(len(ind.Tree))
		start := FindSubtreeBounds(ind.Tree, cut)

		// Generate a single terminal (0 depth forces a terminal)
		newTerminal := GenerateTree(0, alpha, params)

		// Replace the subtree at the cut point with the new terminal
		newTree := append(symbolic.Postfix{}, ind.Tree[:start]...)
		newTree = append(newTree, newTerminal...)
		newTree = append(newTree, ind.Tree[cut+1:]...)

		ind.Tree = newTree
	*/
}
