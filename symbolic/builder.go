package symbolic

// Constant creates a new constant numeric node
func Const(value float64) Node {
	return &ConstantNode{Value: value}
}

// Variable creates a new named input variable node
func Var(name string) Node {
	return &InputNode{Name: name}
}

// NewUnary creates a new unary operator node
func NewUnary(op UnaryOp, input Node) Node {
	return &UnaryNode{
		Op:    op,
		Input: input,
	}
}

// NewBinary creates a new binary operator node
func NewBinary(op BinaryOp, left, right Node) Node {
	return &BinaryNode{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

// --- Convenience Builders for Math Operations ---

func AddNode(left, right Node) Node { return NewBinary(Add, left, right) }
func SubNode(left, right Node) Node { return NewBinary(Sub, left, right) }
func MulNode(left, right Node) Node { return NewBinary(Mul, left, right) }
func DivNode(left, right Node) Node { return NewBinary(Div, left, right) }
func PowNode(left, right Node) Node { return NewBinary(Pow, left, right) }

func SinNode(input Node) Node { return NewUnary(Sin, input) }
func CosNode(input Node) Node { return NewUnary(Cos, input) }
func ExpNode(input Node) Node { return NewUnary(Exp, input) }
func LogNode(input Node) Node { return NewUnary(Log, input) }
