package symbolic

import "fmt"

type Node interface {
	Evaluate() float64
	NumChildren() int
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

func nodeToString(n Node) string {
	if n == nil {
		return ""
	}
	switch node := n.(type) {
	case *InputNode:
		return node.Name
	case *ConstantNode:
		return fmt.Sprintf("%0.2e", node.Value)
	case *UnaryNode:
		return fmt.Sprintf("%s(%s)", node.Op.String(), nodeToString(node.Input))
	case *BinaryNode:
		return fmt.Sprintf("(%s %s %s)", nodeToString(node.Left), node.Op.String(), nodeToString(node.Right))
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
