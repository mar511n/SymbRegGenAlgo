package symbolic

import (
	"math"
	"testing"
)

func TestTreeEvaluationAndString(t *testing.T) {
	// Create tree: sin(x) + (2 * y)
	x := &InputNode{Name: "x", Value: math.Pi / 2} // sin(pi/2) = 1
	y := &InputNode{Name: "y", Value: 3.5}
	two := &ConstantNode{Value: 2.0}

	sinX := &UnaryNode{Op: Sin, Input: x}
	mul := &BinaryNode{Op: Mul, Left: two, Right: y}
	add := &BinaryNode{Op: Add, Left: sinX, Right: mul}

	tree := &Tree{Root: add}

	expectedResult := 1.0 + (2.0 * 3.5) // = 8.0
	if got := tree.Evaluate(); got != expectedResult {
		t.Errorf("Evaluate() = %v, expected %v", got, expectedResult)
	}

	expectedStr := "(sin(x) + (2 * y))"
	if got := tree.String(); got != expectedStr {
		t.Errorf("String() = %v, expected %v", got, expectedStr)
	}

	expectedNodes := 6 // 6 nodes: x, sin, 2, y, mul, add
	if got := tree.NumNodes(); got != expectedNodes {
		t.Errorf("NumNodes() = %v, expected %v", got, expectedNodes)
	}

	expectedDepth := 3 // add is 1, mul is 2, y is 3
	if got := tree.Depth(); got != expectedDepth {
		t.Errorf("Depth() = %v, expected %v", got, expectedDepth)
	}
}

func TestPostfixConversion(t *testing.T) {
	// Original tree: x^2 + 5
	x := &InputNode{Name: "x"}
	two := &ConstantNode{Value: 2.0}
	pow := &BinaryNode{Op: Pow, Left: x, Right: two}
	five := &ConstantNode{Value: 5.0}
	add := &BinaryNode{Op: Add, Left: pow, Right: five}

	originalTree := &Tree{Root: add}
	originalStr := originalTree.String()

	postfix := originalTree.ToPostfix()
	newTree, err := postfix.ToTree()
	if err != nil {
		t.Fatalf("ToTree error: %v", err)
	}

	newStr := newTree.String()
	if newStr != originalStr {
		t.Errorf("Postfix-to-Tree conversion changed structure: got %s, expected %s", newStr, originalStr)
	}
}

func TestParsePostfix(t *testing.T) {
	tests := []struct {
		input    string
		expected string // We will parse it and then convert back to string and see if it equals input, and also evaluated with a tree
		isValid  bool
	}{
		{"x0 exp atan", "x0 exp atan", true},
		{"5.3 x0 exp max atan", "5.3 x0 exp max atan", true},
		{"x1 x2 + 3 *", "x1 x2 + 3 *", true},
		{"sin", "", false},  // not enough arguments
		{"x0 +", "", false}, // not enough arguments
		{"", "", false},     // empty
		{"x 2 * sin", "x 2 * sin", true},
	}

	for _, tt := range tests {
		p, err := ParsePostfix(tt.input)
		if tt.isValid {
			if err != nil {
				t.Errorf("ParsePostfix(%q) returned unexpected error: %v", tt.input, err)
				continue
			}
			if got := p.String(); got != tt.expected {
				t.Errorf("ParsePostfix(%q).String() = %q, expected %q", tt.input, got, tt.expected)
			}

			// Also ensure it can build a tree properly
			_, errTree := p.ToTree()
			if errTree != nil {
				t.Errorf("ParsePostfix(%q).ToTree() returned unexpected error: %v", tt.input, errTree)
			}
		} else {
			if err == nil {
				t.Errorf("ParsePostfix(%q) expected error but got none", tt.input)
			}
		}
	}
}

func TestPostfixConversionOld(t *testing.T) {
	// Original tree: x^2 + 5
	x := &InputNode{Name: "x"}
	two := &ConstantNode{Value: 2.0}
	pow := &BinaryNode{Op: Pow, Left: x, Right: two}
	five := &ConstantNode{Value: 5.0}
	add := &BinaryNode{Op: Add, Left: pow, Right: five}

	originalTree := &Tree{Root: add}
	originalStr := originalTree.String()

	// 1. Convert Tree -> Postfix
	postfix := originalTree.ToPostfix()
	if len(postfix) != 5 {
		t.Fatalf("Expected postfix length 5, got %d", len(postfix))
	}

	// 2. Check Postfix string representation (RPN)
	expectedPostfixStr := "x 2 ^ 5 +"
	if got := postfix.String(); got != expectedPostfixStr {
		t.Errorf("Postfix.String() = %q, expected %q", got, expectedPostfixStr)
	}

	// 3. Convert Postfix -> Tree
	reconstructedTree, err := postfix.ToTree()
	if err != nil {
		t.Fatalf("ToTree() failed: %v", err)
	}

	// 4. Verify reconstruction via string representation
	reconstructedStr := reconstructedTree.String()
	if reconstructedStr != originalStr {
		t.Errorf("Reconstructed tree structure mismatch.\nOriginal: %s\nReconstructed: %s", originalStr, reconstructedStr)
	}
}

func TestInvalidPostfix(t *testing.T) {
	tests := []struct {
		name    string
		postfix Postfix
	}{
		{"Empty slice", Postfix{}},
		{
			"Missing binary operand",
			Postfix{
				{Type: TokenTypeConstant, Value: 1},
				{Type: TokenTypeBinary, BinaryOp: Add},
			},
		},
		{
			"Missing unary operand",
			Postfix{
				{Type: TokenTypeUnary, UnaryOp: Sin},
			},
		},
		{
			"Too many operands left on stack",
			Postfix{
				{Type: TokenTypeConstant, Value: 1},
				{Type: TokenTypeConstant, Value: 2},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := tc.postfix.ToTree()
			if err == nil {
				t.Errorf("Expected error for %s, but got valid tree: %v", tc.name, tree.String())
			}
		})
	}
}
