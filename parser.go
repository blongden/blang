package main

import (
	"errors"
	"fmt"
)

type Tokens struct {
	tokens []Token
	index  int
}

func (t *Tokens) peek() *Token {
	if t.index >= len(t.tokens) {
		return nil
	}
	return &t.tokens[t.index]
}

func (t *Tokens) consume() *Token {
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
	NodePrintln
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

func (s *StatementSequence) append(node *Node) {
	s.statements = append(s.statements, *node)
}

func (t *Tokens) parse_term() *Node {
	if t.peek() == nil {
		return nil
	}

	switch t.peek().token_type {
	case Int:
		return &Node{node_type: NodeIntLiteral, value: t.consume().value}
	case Identifier:
		return &Node{node_type: NodeIdentifier, value: t.consume().value}
	case Lparen:
		t.consume()
		expr := t.parse_expr(0)
		if t.peek() != nil && t.peek().token_type != Rparen {
			panic("Expected ')'")
		}
		t.consume()
		return expr
	default:
		return nil
	}
}

func get_operator_prec(op TokenType) *int {
	var prec int
	switch op {
	case Plus, Minus:
		prec = 0
	case Star, Fslash:
		prec = 1
	default:
		return nil
	}
	return &prec
}

func (t *Tokens) parse_expr(min_prec int) *Node {
	expr := t.parse_term()

	// Future me: read this for an explaination on how this works https://eli.thegreenplace.net/2012/08/02/parsing-expressions-by-precedence-climbing
	for {
		tok := t.peek()
		if tok == nil {
			break
		}

		prec := get_operator_prec(tok.token_type)
		if prec == nil || *prec < min_prec {
			break
		}
		op := t.consume()
		rhs := t.parse_expr(min_prec + 1)
		if rhs == nil {
			panic("unable to parse expression")
		}
		expr2 := Node{lhs: expr, rhs: rhs}
		if op.token_type == Plus {
			expr2.node_type = NodeAdd
		} else if op.token_type == Minus {
			expr2.node_type = NodeSub
		} else if op.token_type == Star {
			expr2.node_type = NodeMulti
		} else if op.token_type == Fslash {
			expr2.node_type = NodeDiv
		} else {
			panic(fmt.Sprintf("Unreachable, this should not happen (see prec check above): token type %d", op.token_type))
		}
		expr = &expr2
	}
	return expr
}

func (t *Tokens) parse_stmt() (*Node, error) {
	if t.peek() == nil {
		return nil, errors.New("no more tokens left")
	}

	switch t.peek().token_type {
	case Exit:
		t.consume()
		var lhs *Node
		if lhs = t.parse_expr(0); lhs == nil {
			lhs = &Node{node_type: NodeIntLiteral, value: "0"}
		}
		return &Node{node_type: NodeExit, lhs: lhs}, nil

	case Let:
		t.consume() // let
		if t.peek() != nil && t.peek().token_type != Identifier {
			panic("Expected identifier")
		}
		lhs := Node{node_type: NodeIdentifier, value: t.consume().value} // x
		if t.peek() != nil && t.peek().token_type != Assign {
			panic("Expected '='")
		}
		t.consume()            // =
		rhs := t.parse_expr(0) // 69
		if rhs == nil {
			panic("Expected expression")
		}
		return &Node{node_type: NodeLet, lhs: &lhs, rhs: rhs}, nil

	case Lcurly:
		stmts := t.parse_scope()
		return &Node{node_type: NodeScope, stmts: stmts}, nil

	case If:
		t.consume()
		lhs := t.parse_expr(0)
		stmts := t.parse_scope()
		return &Node{node_type: NodeIf, lhs: lhs, stmts: stmts}, nil
	case Println:
		t.consume()
		lhs := t.parse_expr(0)
		return &Node{node_type: NodePrintln, lhs: lhs}, nil
	default:
		return nil, errors.New("Unknown statement, " + fmt.Sprint(t.peek().token_type))
	}
}

func (t *Tokens) parse_scope() *StatementSequence {
	if t.peek() == nil || t.peek().token_type != Lcurly {
		panic("Expected '{'")
	}

	t.consume()
	stmts := StatementSequence{}
	for {
		stmt, _ := t.parse_stmt()
		if stmt == nil {
			break
		}
		stmts.append(stmt)
	}
	if t.peek() == nil || t.peek().token_type != Rcurly {
		panic("Expected '}'")
	}
	t.consume()

	return &stmts
}

func (t *Tokens) parse() *StatementSequence {
	stmts := StatementSequence{}

	for {
		if t.peek() == nil {
			break
		}
		stmt, err := t.parse_stmt()
		if stmt == nil {
			panic("Unable to parse statement: " + err.Error())
		}
		stmts.append(stmt)
	}

	return &stmts
}
