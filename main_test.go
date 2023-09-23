package main

import (
	"strconv"
	"testing"
)

func TestExitStatementDefaultsToZero(t *testing.T) {
	tokens := tokenise([]byte("exit"))
	if tokens.peek().token_type != Exit {
		t.Errorf("exit does not generate exit token")
	}

	node, _ := tokens.parse_stmt()

	if node.lhs == nil {
		t.Errorf("exit node has no parameter")
	}

	if node.lhs.node_type != NodeIntLiteral || node.lhs.value != "0" {
		t.Errorf("exit node parameter should default to 0")
	}
}

func TestExitStatementUsesArgument(t *testing.T) {
	tokens := tokenise([]byte("exit 1"))
	if tokens.peek().token_type != Exit {
		t.Errorf("exit does not generate exit token")
	}

	node, _ := tokens.parse_stmt()

	if node.lhs == nil {
		t.Errorf("exit node has no parameter")
	}

	if node.lhs.node_type != NodeIntLiteral || node.lhs.value != "1" {
		t.Errorf("exit node parameter should be set to 1")
	}
}

func evaluateExpr(node *Node) int {
	if node.node_type == NodeIntLiteral {
		i, _ := strconv.Atoi(node.value)
		return i
	} else if node.node_type == NodeAdd {
		return evaluateExpr(node.lhs) + evaluateExpr(node.rhs)
	} else if node.node_type == NodeSub {
		return evaluateExpr(node.lhs) - evaluateExpr(node.rhs)
	} else if node.node_type == NodeMulti {
		return evaluateExpr(node.lhs) * evaluateExpr(node.rhs)
	} else if node.node_type == NodeDiv {
		return evaluateExpr(node.lhs) / evaluateExpr(node.rhs)
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
		tokens := tokenise([]byte(test.expr))
		root := tokens.parse_expr(0)
		result := evaluateExpr(root)
		if result != test.expected {
			t.Errorf("answer incorrect for expression (" + test.expr + "): " + strconv.Itoa(result))
		}
	}
}

func TestLetAssignsVar(t *testing.T) {
	tokens := tokenise([]byte("let x = 5"))
	if tokens.peek().token_type != Let {
		t.Errorf("exit does not generate exit token")
	}

	// need to find a way of testing generation
}
