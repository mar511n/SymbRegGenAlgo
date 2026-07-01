package symbolic

import "math"

type BinaryOp int

const (
	Add BinaryOp = iota
	Sub
	Mul
	Div
	Pow
	Max
	Min
	Mod
)

func (op BinaryOp) String() string {
	switch op {
	case Add:
		return "+"
	case Sub:
		return "-"
	case Mul:
		return "*"
	case Div:
		return "/"
	case Pow:
		return "^"
	case Max:
		return "max"
	case Min:
		return "min"
	case Mod:
		return "mod"
	default:
		return "unknown"
	}
}

var BasicBinaryOps = map[BinaryOp]func(float64, float64) float64{
	Add: func(a, b float64) float64 { return a + b },
	Sub: func(a, b float64) float64 { return a - b },
	Mul: func(a, b float64) float64 { return a * b },
	Div: func(a, b float64) float64 { return a / b },
	Pow: func(a, b float64) float64 { return math.Pow(a, b) },
	Max: func(a, b float64) float64 { return math.Max(a, b) },
	Min: func(a, b float64) float64 { return math.Min(a, b) },
	Mod: func(a, b float64) float64 { return math.Mod(a, b) },
}

type UnaryOp int

const (
	Sin UnaryOp = iota
	Cos
	Exp
	Log
	Abs
	Floor
	Asin
	Acos
	Atan
)

func (op UnaryOp) String() string {
	switch op {
	case Sin:
		return "sin"
	case Cos:
		return "cos"
	case Exp:
		return "exp"
	case Log:
		return "log"
	case Abs:
		return "abs"
	case Floor:
		return "floor"
	case Asin:
		return "asin"
	case Acos:
		return "acos"
	case Atan:
		return "atan"
	default:
		return "unknown"
	}
}

var BasicUnaryOps = map[UnaryOp]func(float64) float64{
	Sin:   func(a float64) float64 { return math.Sin(a) },
	Cos:   func(a float64) float64 { return math.Cos(a) },
	Exp:   func(a float64) float64 { return math.Exp(a) },
	Log:   func(a float64) float64 { return math.Log(a) },
	Abs:   func(a float64) float64 { return math.Abs(a) },
	Floor: func(a float64) float64 { return math.Floor(a) },
	Asin:  func(a float64) float64 { return math.Asin(a) },
	Acos:  func(a float64) float64 { return math.Acos(a) },
	Atan:  func(a float64) float64 { return math.Atan(a) },
}

type Tree struct {
	Root Node
}

func (t *Tree) RemoveNaNInf() {
	if t != nil && t.Root != nil {
		t.Root = removeNaNInf(t.Root)
	}
}

func (t *Tree) Evaluate() float64 {
	return t.Root.Evaluate()
}

func (t *Tree) Simplify() {
	if t != nil && t.Root != nil {
		t.Root = Simplify(t.Root)
		t.Root = removeNaNInf(t.Root)
	}
}

func (t *Tree) NumNodes() int {
	return countNodes(t.Root)
}

func (t *Tree) Depth() int {
	return nodeDepth(t.Root)
}

func (t *Tree) String() string {
	return nodeToString(t.Root, "%0.2e")
}

func (t *Tree) Stringfmt(floatFormat string) string {
	return nodeToString(t.Root, floatFormat)
}

func (t *Tree) ToPostfix() Postfix {
	if t == nil || t.Root == nil {
		return nil
	}
	return nodeToPostfix(t.Root)
}
