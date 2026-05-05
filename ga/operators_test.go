package ga

import (
	"fmt"
	"testing"
)

func MakeTestSetup() (*Alphabet, *GeneratorParams, *Individual) {
	rnd.Seed(0)
	params := DefaultGeneratorParams()
	alph := DefaultAlphabet([]string{"x", "y"})
	tree := GenerateTree(5, alph, params)
	return alph, params, &Individual{Tree: tree}
}

func TestGrowLeaf(t *testing.T) {
	alph, params, ind := MakeTestSetup()
	tree, _ := ind.Tree.ToTree()
	fmt.Printf("Original tree: %s\n", tree.String())
	fmt.Printf("Original postfix: %s\n", ind.Tree.String())

	for i := range 5 {
		indCopy := ind.Copy()
		GrowLeafMutation(indCopy, alph, 5, params)
		tree, _ := indCopy.Tree.ToTree()
		fmt.Printf("After GrowLeafMutation %d: %s\n", i+1, tree.String())
		fmt.Printf("Postfix: %s\n", indCopy.Tree.String())
	}
	t.Fail()
}

func TestGrowRoot(t *testing.T) {
	alph, params, ind := MakeTestSetup()
	tree, _ := ind.Tree.ToTree()
	fmt.Printf("Original tree: %s\n", tree.String())
	fmt.Printf("Original postfix: %s\n", ind.Tree.String())

	for i := range 5 {
		indCopy := ind.Copy()
		GrowRootMutation(indCopy, alph, 5, params)
		tree, _ := indCopy.Tree.ToTree()
		fmt.Printf("After GrowRootMutation %d: %s\n", i+1, tree.String())
		fmt.Printf("Postfix: %s\n", indCopy.Tree.String())
	}
	t.Fail()
}

func TestRemoveUnary(t *testing.T) {
	_, _, ind := MakeTestSetup()
	tree, _ := ind.Tree.ToTree()
	fmt.Printf("Original tree: %s\n", tree.String())
	fmt.Printf("Original postfix: %s\n", ind.Tree.String())

	for i := range 5 {
		indCopy := ind.Copy()
		RemoveUnaryMutation(indCopy)
		tree, _ := indCopy.Tree.ToTree()
		fmt.Printf("After RemoveUnaryMutation %d: %s\n", i+1, tree.String())
		fmt.Printf("Postfix: %s\n", indCopy.Tree.String())
	}
	t.Fail()
}

func TestShrink(t *testing.T) {
	alph, params, ind := MakeTestSetup()
	tree, _ := ind.Tree.ToTree()
	fmt.Printf("Original tree: %s\n", tree.String())
	fmt.Printf("Original postfix: %s\n", ind.Tree.String())

	for i := range 5 {
		indCopy := ind.Copy()
		ShrinkMutation(indCopy, alph, 5, params)
		tree, _ := indCopy.Tree.ToTree()
		fmt.Printf("After ShrinkMutation %d: %s\n", i+1, tree.String())
		fmt.Printf("Postfix: %s\n", indCopy.Tree.String())
	}
	t.Fail()
}
