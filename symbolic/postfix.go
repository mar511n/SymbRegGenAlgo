package symbolic

import (
	"fmt"
	"strconv"
	"strings"
)

type TokenType int

const (
	TokenTypeConstant TokenType = iota
	TokenTypeVariable
	TokenTypeUnary
	TokenTypeBinary
)

// Token represents a single elements in a flat postfix array.
type Token struct {
	Type     TokenType
	BinaryOp BinaryOp
	UnaryOp  UnaryOp
	Name     string
	Value    float64
}

// Postfix represents a mathematical expression tree in Reverse Polish Notation (from bottom-up).
type Postfix []Token

func (p Postfix) TokenDiff(other Postfix) (d float64) {
	d = 0
	smallP := p
	largerP := other
	if len(other) < len(p) {
		smallP = other
		largerP = p
	}

	// calculate number of tokens present in smallP but not in largerP
	tokenCount := make(map[Token]int)
	for _, token := range largerP {
		tokenCount[token]++
	}
	for _, token := range smallP {
		if tokenCount[token] > 0 {
			tokenCount[token]--
		} else {
			d += 1
		}
	}
	return
}

// ToTree builds a tree from the postfix array using a standard stack approach.
func (p Postfix) ToTree() (*Tree, error) {
	if len(p) == 0 {
		return nil, fmt.Errorf("empty postfix expression")
	}

	stack := make([]Node, 0, len(p))

	for _, token := range p {
		switch token.Type {
		case TokenTypeConstant:
			stack = append(stack, &ConstantNode{Value: token.Value})
		case TokenTypeVariable:
			stack = append(stack, &InputNode{Name: token.Name})
		case TokenTypeUnary:
			if len(stack) < 1 {
				return nil, fmt.Errorf("not enough arguments for unary operator")
			}
			input := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			stack = append(stack, &UnaryNode{Op: token.UnaryOp, Input: input})
		case TokenTypeBinary:
			if len(stack) < 2 {
				return nil, fmt.Errorf("not enough arguments for binary operator")
			}
			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, &BinaryNode{Op: token.BinaryOp, Left: left, Right: right})
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("invalid postfix expression: ended with %d items on the stack", len(stack))
	}

	return &Tree{Root: stack[0]}, nil
}

func (p Postfix) Simplify() Postfix {
	tree, err := p.ToTree()
	if err != nil {
		return p // Return original if invalid
	}
	tree.Simplify()
	return tree.ToPostfix()
}

func nodeToPostfix(n Node) Postfix {
	var p Postfix
	var traverse func(node Node)

	traverse = func(node Node) {
		if node == nil {
			return
		}
		switch val := node.(type) {
		case *ConstantNode:
			p = append(p, Token{Type: TokenTypeConstant, Value: val.Value})
		case *InputNode:
			p = append(p, Token{Type: TokenTypeVariable, Name: val.Name})
		case *UnaryNode:
			traverse(val.Input)
			p = append(p, Token{Type: TokenTypeUnary, UnaryOp: val.Op})
		case *BinaryNode:
			traverse(val.Left)
			traverse(val.Right)
			p = append(p, Token{Type: TokenTypeBinary, BinaryOp: val.Op})
		}
	}

	traverse(n)
	return p
}

// String creates a human readable version of the postfix array (e.g. "x 2 * sin")
func (p Postfix) String() string {
	var parts []string
	for _, token := range p {
		switch token.Type {
		case TokenTypeConstant:
			parts = append(parts, strconv.FormatFloat(token.Value, 'g', -1, 64))
		case TokenTypeVariable:
			parts = append(parts, token.Name)
		case TokenTypeUnary:
			parts = append(parts, token.UnaryOp.String())
		case TokenTypeBinary:
			parts = append(parts, token.BinaryOp.String())
		}
	}
	return strings.Join(parts, " ")
}

// ParsePostfix parses a postfix string (e.g., "x0 exp atan") into a Postfix token array.
func ParsePostfix(s string) (Postfix, error) {
	if strings.TrimSpace(s) == "" {
		return nil, fmt.Errorf("empty postfix string")
	}

	parts := strings.Fields(s)

	// Create lookup maps for fast matching
	strToBinary := map[string]BinaryOp{
		"+":   Add,
		"-":   Sub,
		"*":   Mul,
		"/":   Div,
		"^":   Pow,
		"max": Max,
		"min": Min,
		"mod": Mod,
	}

	strToUnary := map[string]UnaryOp{
		"sin":   Sin,
		"cos":   Cos,
		"exp":   Exp,
		"log":   Log,
		"abs":   Abs,
		"floor": Floor,
		"asin":  Asin,
		"acos":  Acos,
		"atan":  Atan,
	}

	var p Postfix
	for _, part := range parts {
		if val, err := strconv.ParseFloat(part, 64); err == nil {
			p = append(p, Token{Type: TokenTypeConstant, Value: val})
			continue
		}

		if op, ok := strToBinary[part]; ok {
			p = append(p, Token{Type: TokenTypeBinary, BinaryOp: op})
			continue
		}

		if op, ok := strToUnary[part]; ok {
			p = append(p, Token{Type: TokenTypeUnary, UnaryOp: op})
			continue
		}

		// Variable fallback
		p = append(p, Token{Type: TokenTypeVariable, Name: part})
	}

	// Optional runtime check if sequence is valid by making an ephemeral tree
	if _, err := p.ToTree(); err != nil {
		return nil, fmt.Errorf("invalid postfix expression structure: %w", err)
	}

	return p, nil
}
