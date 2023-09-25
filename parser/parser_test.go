package parser

import (
	"strconv"
	"testing"

	"longden.me/blang/tokeniser"
)

func TestExitStatementDefaultsToZero(t *testing.T) {
	tokens, _ := tokeniser.Tokenise([]byte("exit"))
	p := Parser{Tokens: tokens}
	if p.peek().Type != tokeniser.Exit {
		t.Errorf("exit does not generate exit token")
	}

	node, _ := p.parse_stmt()
	if node.Lhs == nil {
		t.Errorf("exit node has no parameter")
	}

	if node.Lhs.Type != NodeIntLiteral || node.Lhs.Value != "0" {
		t.Errorf("exit node parameter should default to 0")
	}
}

func TestExitStatementUsesArgument(t *testing.T) {
	tokens, _ := tokeniser.Tokenise([]byte("exit 1"))
	parser := Parser{Tokens: tokens}
	if parser.peek().Type != tokeniser.Exit {
		t.Errorf("exit does not generate exit token")
	}

	node, _ := parser.parse_stmt()

	if node.Lhs == nil {
		t.Errorf("exit node has no parameter")
	}

	if node.Lhs.Type != NodeIntLiteral || node.Lhs.Value != "1" {
		t.Errorf("exit node parameter should be set to 1")
	}
}

func evaluateExpr(node *Node) int {
	if node.Type == NodeIntLiteral {
		i, _ := strconv.Atoi(node.Value)
		return i
	} else if node.Type == NodeAdd {
		return evaluateExpr(node.Lhs) + evaluateExpr(node.Rhs)
	} else if node.Type == NodeSub {
		return evaluateExpr(node.Lhs) - evaluateExpr(node.Rhs)
	} else if node.Type == NodeMulti {
		return evaluateExpr(node.Lhs) * evaluateExpr(node.Rhs)
	} else if node.Type == NodeDiv {
		return evaluateExpr(node.Lhs) / evaluateExpr(node.Rhs)
	}
	return 0
}

type exprTest struct {
	expr     string
	expected int
}

var exprTests = []exprTest{
	{"6 / 3", 2},
	{"2 + 3 * 3 + 2", 13},
	{"3 * 3 + 2 + 2", 13},
	{"2 + 2 + 3 * 3", 13},
	{"1 + 2 + 6 / 3 - 1", 4},
	{"(1 + 4) * 8 / 2 - 3", 17},
	{"1 + 4 * 8 / (2 + 3)", 7}, // rounds to int at the moment
}

func TestExprPrecedenceClimbingMulti(t *testing.T) {
	for _, test := range exprTests {
		tokens, _ := tokeniser.Tokenise([]byte(test.expr))
		parser := Parser{Tokens: tokens}
		root, _ := parser.parse_expr(0)
		result := evaluateExpr(root)
		if result != test.expected {
			t.Errorf("answer incorrect for expression (" + test.expr + "): " + strconv.Itoa(result))
		}
	}
}

func TestExprInvalid(t *testing.T) {
	tokens, _ := tokeniser.Tokenise([]byte("let x = 2 +"))
	parser := Parser{Tokens: tokens}
	_, err := parser.parse_stmt()
	if err == nil {
		t.Errorf("expected invalid expression error")
	}
}

func TestExprParenInvalid(t *testing.T) {
	tokens, _ := tokeniser.Tokenise([]byte("let x = 2 + (2"))
	p := Parser{Tokens: tokens}
	_, err := p.parse_stmt()
	if err == nil {
		t.Errorf("expected invalid expression error")
	}
}
