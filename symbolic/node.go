package symbolic

import (
	"fmt"
	"math"
)

type Node interface {
	Evaluate() float64
	NumChildren() int
}

func removeNaNInf(n Node) Node {
	if n == nil {
		return nil
	}
	// check if current node is a binary node.
	// If so, check if either child is ConstantNode with NaN or Inf value.
	// If both are NaN or Inf, returns a ConstantNode with value 0.
	// If only one is NaN or Inf, returns the other child.
	// If no children are NaN or Inf, check both children and return self.
	switch node := n.(type) {
	case *BinaryNode:
		leftConst, leftIsConst := node.Left.(*ConstantNode)
		rightConst, rightIsConst := node.Right.(*ConstantNode)
		leftNaNInf := leftIsConst && isNaNOrInf(leftConst.Value)
		rightNaNInf := rightIsConst && isNaNOrInf(rightConst.Value)

		if leftNaNInf && rightNaNInf {
			return &ConstantNode{Value: 0}
		} else if leftNaNInf {
			return removeNaNInf(node.Right)
		} else if rightNaNInf {
			return removeNaNInf(node.Left)
		}
		node.Left = removeNaNInf(node.Left)
		node.Right = removeNaNInf(node.Right)
	case *UnaryNode:
		node.Input = removeNaNInf(node.Input)
	}
	return n
}

func isNaNOrInf(val float64) bool {
	return math.IsNaN(val) || math.IsInf(val, 0)
}

func countNodes(n Node) int {
	if n == nil {
		return 0
	}
	count := 1
	switch child := n.(type) {
	case *UnaryNode:
		count += countNodes(child.Input)
	case *BinaryNode:
		count += countNodes(child.Left)
		count += countNodes(child.Right)
	}
	return count
}

func nodeDepth(n Node) int {
	if n == nil {
		return 0
	}
	switch child := n.(type) {
	case *UnaryNode:
		return 1 + nodeDepth(child.Input)
	case *BinaryNode:
		leftDepth := nodeDepth(child.Left)
		rightDepth := nodeDepth(child.Right)
		if leftDepth > rightDepth {
			return 1 + leftDepth
		}
		return 1 + rightDepth
	default:
		return 1
	}
}

func nodeToString(n Node, floatFormat string) string {
	if n == nil {
		return ""
	}
	switch node := n.(type) {
	case *InputNode:
		return node.Name
	case *ConstantNode:
		return fmt.Sprintf(floatFormat, node.Value)
	case *UnaryNode:
		return fmt.Sprintf("%s(%s)", node.Op.String(), nodeToString(node.Input, floatFormat))
	case *BinaryNode:
		return fmt.Sprintf("(%s %s %s)", nodeToString(node.Left, floatFormat), node.Op.String(), nodeToString(node.Right, floatFormat))
	}
	return "unknown"
}

type InputNode struct {
	Name  string
	Value float64
}

func (n *InputNode) Evaluate() float64 {
	return n.Value
}

func (n *InputNode) NumChildren() int {
	return 0
}

func Simplify(n Node) Node {
	if n == nil {
		return nil
	}

	switch node := n.(type) {
	case *UnaryNode:
		node.Input = Simplify(node.Input)
		if _, ok := node.Input.(*ConstantNode); ok {
			return &ConstantNode{Value: node.Evaluate()}
		}
	case *BinaryNode:
		node.Left = Simplify(node.Left)
		node.Right = Simplify(node.Right)
		_, leftIsConst := node.Left.(*ConstantNode)
		_, rightIsConst := node.Right.(*ConstantNode)
		if leftIsConst && rightIsConst {
			return &ConstantNode{Value: node.Evaluate()}
		}
		_, leftIsVar := node.Left.(*InputNode)
		_, rightIsVar := node.Right.(*InputNode)
		// list of possible simplifications
		/*
			x + 0 = x, 0 + x = x
			x - 0 = x
			x * 1 = x, 1 * x = x
			x / 1 = x
			x * 0 = 0, 0 * x = 0
			0 / x = 0
			x / x = 1
			x - x = 0
		*/
		switch node.Op {
		case Add:
			if leftIsConst && node.Left.(*ConstantNode).Value == 0 {
				return node.Right
			}
			if rightIsConst && node.Right.(*ConstantNode).Value == 0 {
				return node.Left
			}
		case Sub:
			if rightIsConst && node.Right.(*ConstantNode).Value == 0 {
				return node.Left
			}
			if leftIsVar && rightIsVar && node.Left.(*InputNode).Name == node.Right.(*InputNode).Name {
				return &ConstantNode{Value: 0}
			}
		case Mul:
			if leftIsConst && node.Left.(*ConstantNode).Value == 1 {
				return node.Right
			}
			if rightIsConst && node.Right.(*ConstantNode).Value == 1 {
				return node.Left
			}
			if leftIsConst && node.Left.(*ConstantNode).Value == 0 {
				return &ConstantNode{Value: 0}
			}
			if rightIsConst && node.Right.(*ConstantNode).Value == 0 {
				return &ConstantNode{Value: 0}
			}
		case Div:
			if rightIsConst && node.Right.(*ConstantNode).Value == 1 {
				return node.Left
			}
			if leftIsConst && node.Left.(*ConstantNode).Value == 0 {
				return &ConstantNode{Value: 0}
			}
			if leftIsVar && rightIsVar && node.Left.(*InputNode).Name == node.Right.(*InputNode).Name {
				return &ConstantNode{Value: 1}
			}
		}
	}

	return n
}

type ConstantNode struct {
	Value float64
}

func (n *ConstantNode) Evaluate() float64 {
	return n.Value
}

func (n *ConstantNode) NumChildren() int {
	return 0
}

type UnaryNode struct {
	Input Node
	Op    UnaryOp
}

func (n *UnaryNode) Evaluate() float64 {
	return BasicUnaryOps[n.Op](n.Input.Evaluate())
}

func (n *UnaryNode) NumChildren() int {
	return 1
}

type BinaryNode struct {
	Left  Node
	Right Node
	Op    BinaryOp
}

func (n *BinaryNode) Evaluate() float64 {
	return BasicBinaryOps[n.Op](n.Left.Evaluate(), n.Right.Evaluate())
}

func (n *BinaryNode) NumChildren() int {
	return 2
}
