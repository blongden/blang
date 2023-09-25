package main

import (
	"errors"
	"fmt"

	"longden.me/blang/tokeniser"
)

type Parser struct {
	tokens []tokeniser.Token
	index  int
}

func (t *Parser) peek() *tokeniser.Token {
	if t.index >= len(t.tokens) {
		return nil
	}
	return &t.tokens[t.index]
}

func (t *Parser) consume() *tokeniser.Token {
	if t.index >= len(t.tokens) {
		return nil
	}
	tok := &t.tokens[t.index]
	t.index++
	return tok
}

type NodeType int

const (
	Sequence NodeType = iota
	NodeExit
	NodeIntLiteral
	NodeIdentifier
	NodeLet
	NodeAdd
	NodeSub
	NodeMulti
	NodeDiv
	NodeScope
	NodeIf
	NodeLt
	NodeGt
	NodeEq
	NodeAssign
	NodeFor
)

type StatementSequence struct {
	statements []Node
}

type Node struct {
	node_type NodeType
	value     string
	lhs       *Node
	rhs       *Node
	stmts     *StatementSequence
}

func parse_error(message string, token *tokeniser.Token) error {
	return fmt.Errorf(message+" at line %d, col %d", token.Line, token.Col)
}

func (s *StatementSequence) append(node *Node) {
	s.statements = append(s.statements, *node)
}

func (t *Parser) parse_term() *Node {
	if t.peek() == nil {
		return nil
	}

	switch t.peek().Type {
	case tokeniser.Int:
		return &Node{node_type: NodeIntLiteral, value: t.consume().Value}
	case tokeniser.Identifier:
		return &Node{node_type: NodeIdentifier, value: t.consume().Value}
	case tokeniser.Lparen:
		t.consume()
		expr := t.parse_expr(0)
		if t.peek() != nil && t.peek().Type != tokeniser.Rparen {
			panic("Expected ')'")
		}
		t.consume()
		return expr
	default:
		return nil
	}
}

func get_operator_prec(op tokeniser.TokenType) *int {
	var prec int
	switch op {
	case tokeniser.Plus, tokeniser.Minus:
		prec = 0
	case tokeniser.Star, tokeniser.Fslash:
		prec = 1
	default:
		return nil
	}
	return &prec
}

func (t *Parser) parse_test() *Node {
	test := t.parse_expr(0)

	tok := t.peek()
	if tok != nil {
		switch tok.Type {
		case tokeniser.Gt:
			t.consume()
			node := Node{node_type: NodeGt, lhs: test, rhs: t.parse_expr(0)}
			test = &node
		case tokeniser.Lt:
			t.consume()
			node := Node{node_type: NodeLt, lhs: test, rhs: t.parse_expr(0)}
			test = &node
		case tokeniser.Eq:
			t.consume()
			node := Node{node_type: NodeEq, lhs: test, rhs: t.parse_expr(0)}
			test = &node
		default:
			node := Node{node_type: NodeGt, lhs: test, rhs: &Node{node_type: NodeIntLiteral, value: "0"}}
			test = &node
		}
	}

	return test
}

func (t *Parser) parse_expr(min_prec int) *Node {
	expr := t.parse_term()

	// Future me: read this for an explaination on how this works https://eli.thegreenplace.net/2012/08/02/parsing-expressions-by-precedence-climbing
	for {
		tok := t.peek()
		if tok == nil {
			break
		}

		prec := get_operator_prec(tok.Type)
		if prec == nil || *prec < min_prec {
			break
		}
		op := t.consume()
		rhs := t.parse_expr(min_prec + 1)
		if rhs == nil {
			panic("unable to parse expression")
		}
		expr2 := Node{lhs: expr, rhs: rhs}
		if op.Type == tokeniser.Plus {
			expr2.node_type = NodeAdd
		} else if op.Type == tokeniser.Minus {
			expr2.node_type = NodeSub
		} else if op.Type == tokeniser.Star {
			expr2.node_type = NodeMulti
		} else if op.Type == tokeniser.Fslash {
			expr2.node_type = NodeDiv
		} else {
			panic(fmt.Sprintf("Unreachable, this should not happen (see prec check above): token type %d", op.Type))
		}
		expr = &expr2
	}
	return expr
}

func (t *Parser) parse_stmt() (*Node, error) {
	if t.peek() == nil {
		return nil, errors.New("no more tokens left")
	}

	switch t.peek().Type {
	case tokeniser.Exit:
		t.consume()
		var lhs *Node
		if lhs = t.parse_expr(0); lhs == nil {
			lhs = &Node{node_type: NodeIntLiteral, value: "0"}
		}
		return &Node{node_type: NodeExit, lhs: lhs}, nil

	case tokeniser.Let:
		c := t.consume() // let
		if t.peek() != nil && t.peek().Type != tokeniser.Identifier {
			return nil, parse_error("expected identifier", c)
		}

		c = t.consume()
		lhs := Node{node_type: NodeIdentifier, value: c.Value} // x
		if t.peek() != nil && t.peek().Type != tokeniser.Assign {
			return nil, parse_error("expected '='", c)
		}

		c = t.consume()        // =
		rhs := t.parse_expr(0) // 69
		if rhs == nil {
			return nil, parse_error("expected expression", c)
		}

		return &Node{node_type: NodeLet, lhs: &lhs, rhs: rhs}, nil

	case tokeniser.Lcurly:
		stmts, _ := t.parse_scope()
		return &Node{node_type: NodeScope, stmts: stmts}, nil

	case tokeniser.If:
		t.consume()

		c := t.peek() // for error reporting
		lhs := t.parse_test()

		if lhs == nil {
			return nil, parse_error("expected test", c)
		}

		stmts, err := t.parse_scope()
		if err != nil {
			return nil, err
		}
		return &Node{node_type: NodeIf, lhs: lhs, stmts: stmts}, nil

	case tokeniser.Identifier:
		lhs := Node{node_type: NodeIdentifier, value: t.consume().Value}
		if t.peek() != nil && t.peek().Type != tokeniser.Assign {
			panic("Expected an assignment; got " + fmt.Sprint(t.peek().Type))
		}
		t.consume()
		return &Node{node_type: NodeAssign, lhs: &lhs, rhs: t.parse_expr(0)}, nil

	case tokeniser.For:
		t.consume()
		lhs := t.parse_test()
		stmts, _ := t.parse_scope()
		return &Node{node_type: NodeFor, lhs: lhs, stmts: stmts}, nil

	default:
		return nil, errors.New("Unknown statement, " + fmt.Sprint(t.peek().Type))
	}
}

func (t *Parser) parse_scope() (*StatementSequence, error) {
	if t.peek() == nil {
		return nil, fmt.Errorf("unexpected end of file")
	}

	if t.peek().Type != tokeniser.Lcurly {
		return nil, parse_error("expected '{'", t.peek())
	}

	c := t.consume()
	stmts := StatementSequence{}
	for {
		stmt, _ := t.parse_stmt()
		if stmt == nil {
			break
		}
		stmts.append(stmt)
	}

	if len(stmts.statements) == 0 {
		return nil, parse_error("at least one statement required in scope", c)
	}

	if t.peek() == nil || t.peek().Type != tokeniser.Rcurly {
		return nil, parse_error("expected '}'", t.peek())
	}
	t.consume()

	return &stmts, nil
}

func (t *Parser) parse() *StatementSequence {
	stmts := StatementSequence{}

	for {
		if t.peek() == nil {
			break
		}
		stmt, err := t.parse_stmt()
		if stmt == nil {
			panic("Parse error: " + err.Error())
		}
		stmts.append(stmt)
	}

	return &stmts
}
