package symbolic

import (
	"testing"
)

func TestSimplify(t *testing.T) {
	// 1. Basic constant expression: 3 * (2 + 1) -> 9
	n1 := &BinaryNode{
		Op:   Mul,
		Left: &ConstantNode{Value: 3},
		Right: &BinaryNode{
			Op:    Add,
			Left:  &ConstantNode{Value: 2},
			Right: &ConstantNode{Value: 1},
		},
	}

	tree := &Tree{Root: n1}
	tree.Simplify()

	if root, ok := tree.Root.(*ConstantNode); !ok {
		t.Errorf("Expected root to be a ConstantNode, got %T", tree.Root)
	} else if root.Value != 9 {
		t.Errorf("Expected simplified value 9, got %v", root.Value)
	}

	// 2. Expression with variable: 3 * (2 + x) -> keep structure, partially simplified?
	// Actually (2 + x) has a variable, so shouldn't simplify.
	n2 := &BinaryNode{
		Op:   Mul,
		Left: &ConstantNode{Value: 3},
		Right: &BinaryNode{
			Op:    Add,
			Left:  &ConstantNode{Value: 2},
			Right: &InputNode{Name: "x"},
		},
	}

	tree2 := &Tree{Root: n2}
	tree2.Simplify()

	if _, ok := tree2.Root.(*BinaryNode); !ok {
		t.Errorf("Expected root to still be a BinaryNode, got %T", tree2.Root)
	}

	// 3. Unary node simplification: abs(-5) -> 5
	n3 := &UnaryNode{
		Op:    Abs,
		Input: &ConstantNode{Value: -5},
	}

	tree3 := &Tree{Root: n3}
	tree3.Simplify()

	if root, ok := tree3.Root.(*ConstantNode); !ok {
		t.Errorf("Expected root to be a ConstantNode, got %T", tree3.Root)
	} else if root.Value != 5 {
		t.Errorf("Expected simplified value 5, got %v", root.Value)
	}

	// 4. Mixed tree: abs(-5) + x -> 5 + x
	n4 := &BinaryNode{
		Op: Add,
		Left: &UnaryNode{
			Op:    Abs,
			Input: &ConstantNode{Value: -5},
		},
		Right: &InputNode{Name: "x"},
	}

	tree4 := &Tree{Root: n4}
	tree4.Simplify()

	root4, ok := tree4.Root.(*BinaryNode)
	if !ok {
		t.Errorf("Expected root to be a BinaryNode, got %T", tree4.Root)
	} else {
		if left, ok := root4.Left.(*ConstantNode); !ok || left.Value != 5 {
			t.Errorf("Expected left to be ConstantNode with value 5")
		}
	}
}
