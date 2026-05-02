package ga

import (
	"marvin/symbreggenalgo/symbolic"
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

// Crossover performs subtree crossover swapping node segments via slice manipulation.
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
		// Point Mutation
		idx := rnd.Intn(len(ind.Tree))
		token := &ind.Tree[idx]
		switch token.Type {
		case symbolic.TokenTypeBinary:
			if len(alpha.BinaryOps) > 0 {
				token.BinaryOp = alpha.BinaryOps[rnd.Intn(len(alpha.BinaryOps))]
			}
		case symbolic.TokenTypeUnary:
			if len(alpha.UnaryOps) > 0 {
				token.UnaryOp = alpha.UnaryOps[rnd.Intn(len(alpha.UnaryOps))]
			}
		case symbolic.TokenTypeConstant:
			token.Value = alpha.MinConst + rnd.Float64()*(alpha.MaxConst-alpha.MinConst)
		case symbolic.TokenTypeVariable:
			if len(alpha.Variables) > 0 {
				token.Name = alpha.Variables[rnd.Intn(len(alpha.Variables))]
			}
		}
	} else if randVal < params.PointMutationProb+params.SubtreeMutationProb {
		// Subtree Mutation
		cut := rnd.Intn(len(ind.Tree))
		start := FindSubtreeBounds(ind.Tree, cut)

		newSubtree := GenerateTree(maxDepth, alpha, params)

		newTree := append(symbolic.Postfix{}, ind.Tree[:start]...)
		newTree = append(newTree, newSubtree...)
		newTree = append(newTree, ind.Tree[cut+1:]...)

		ind.Tree = newTree
	} else {
		// Shrink Mutation
		// Replaces a randomly selected subtree with a randomly generated terminal (depth 0)
		cut := rnd.Intn(len(ind.Tree))
		start := FindSubtreeBounds(ind.Tree, cut)

		// Generate a single terminal (0 depth forces a terminal)
		newTerminal := GenerateTree(0, alpha, params)

		newTree := append(symbolic.Postfix{}, ind.Tree[:start]...)
		newTree = append(newTree, newTerminal...)
		newTree = append(newTree, ind.Tree[cut+1:]...)

		ind.Tree = newTree
	}
	ind.Tree = ind.Tree.Simplify()
}
