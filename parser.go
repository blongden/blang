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

func (t *Parser) parse_term() (*Node, error) {
	tok := t.peek()
	if tok == nil {
		return nil, nil
	}

	switch tok.Type {
	case tokeniser.Int:
		return &Node{node_type: NodeIntLiteral, value: t.consume().Value}, nil
	case tokeniser.Identifier:
		return &Node{node_type: NodeIdentifier, value: t.consume().Value}, nil
	case tokeniser.Lparen:
		t.consume()
		expr, err := t.parse_expr(0)
		if err != nil {
			return nil, err
		}

		if t.peek() == nil {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if t.peek().Type != tokeniser.Rparen {
			return nil, parse_error("expected ')'", t.peek())
		}
		t.consume()
		return expr, nil
	default:
		return nil, nil
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

func (t *Parser) parse_test() (*Node, error) {
	test, err := t.parse_expr(0)

	if err != nil {
		return nil, err
	}

	tok := t.peek()
	if tok != nil {
		switch tok.Type {
		case tokeniser.Gt:
			t.consume()
			rhs, err := t.parse_expr(0)
			if err != nil {
				return nil, err
			}
			node := Node{node_type: NodeGt, lhs: test, rhs: rhs}
			test = &node
		case tokeniser.Lt:
			t.consume()
			rhs, err := t.parse_expr(0)
			if err != nil {
				return nil, err
			}
			node := Node{node_type: NodeLt, lhs: test, rhs: rhs}
			test = &node
		case tokeniser.Eq:
			t.consume()
			rhs, err := t.parse_expr(0)
			if err != nil {
				return nil, err
			}
			node := Node{node_type: NodeEq, lhs: test, rhs: rhs}
			test = &node
		default:
			node := Node{node_type: NodeGt, lhs: test, rhs: &Node{node_type: NodeIntLiteral, value: "0"}}
			test = &node
		}
	}

	return test, nil
}

func (t *Parser) parse_expr(min_prec int) (*Node, error) {
	expr, err := t.parse_term()

	if err != nil {
		return nil, err
	}

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
		rhs, err := t.parse_expr(min_prec + 1)
		if err != nil {
			return nil, err
		}

		if rhs == nil {
			return nil, parse_error("invalid expression", op)
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
	return expr, nil
}

func (t *Parser) parse_stmt() (*Node, error) {
	if t.peek() == nil {
		return nil, errors.New("no more tokens left")
	}

	switch t.peek().Type {
	case tokeniser.Exit:
		t.consume()
		var lhs *Node
		lhs, _ = t.parse_expr(0)
		if lhs == nil {
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

		t.consume()                 // =
		rhs, err := t.parse_expr(0) // 69
		if err != nil {
			return nil, err
		}

		return &Node{node_type: NodeLet, lhs: &lhs, rhs: rhs}, nil

	case tokeniser.Lcurly:
		stmts, _ := t.parse_scope()
		return &Node{node_type: NodeScope, stmts: stmts}, nil

	case tokeniser.If:
		t.consume()

		lhs, err := t.parse_test()

		if err != nil {
			return nil, err
		}

		stmts, err := t.parse_scope()
		if err != nil {
			return nil, err
		}
		return &Node{node_type: NodeIf, lhs: lhs, stmts: stmts}, nil

	case tokeniser.Identifier:
		id := t.consume()
		lhs := Node{node_type: NodeIdentifier, value: id.Value}
		if t.peek() == nil {
			return nil, parse_error("unexpected end of file", id)
		}

		if t.peek().Type != tokeniser.Assign {
			return nil, parse_error(fmt.Sprintf("expected an assignment, got %d", t.peek().Type), t.peek())
		}
		t.consume()
		rhs, err := t.parse_expr(0)
		if err != nil {
			return nil, err
		}
		return &Node{node_type: NodeAssign, lhs: &lhs, rhs: rhs}, nil

	case tokeniser.For:
		t.consume()
		lhs, err := t.parse_test()
		if err != nil {
			return nil, err
		}
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
