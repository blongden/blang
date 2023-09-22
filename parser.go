package main

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
)

type Node struct {
	node_type NodeType
	String    string
	lhs       *Node
	rhs       *Node
}

type StatementSequence struct {
	statements []Node
}

func (s *StatementSequence) append(node *Node) {
	s.statements = append(s.statements, *node)
}

func (t *Tokens) parse_expr() *Node {
	if t.peek() == nil {
		return nil
	}

	switch t.peek().token_type {
	case Int:
		return &Node{node_type: NodeIntLiteral, String: t.consume().value}
	case Identifier:
		return &Node{node_type: NodeIdentifier, String: t.consume().value}
	default:
		return nil
	}
}

func (t *Tokens) parse() *StatementSequence {
	stmts := StatementSequence{}
	for t.peek() != nil {
		switch t.peek().token_type {
		case Exit:
			t.consume()
			var lhs *Node
			if lhs = t.parse_expr(); lhs == nil {
				lhs = &Node{node_type: NodeIntLiteral, String: "0"}
			}
			stmts.append(&Node{node_type: NodeExit, lhs: lhs})

		case Let:
			t.consume() // let
			if t.peek() != nil && t.peek().token_type != Identifier {
				panic("Expected identifier")
			}
			lhs := Node{node_type: NodeIdentifier, String: t.consume().value} // x
			if t.peek() != nil && t.peek().token_type != Eq {
				panic("Expected '='")
			}
			t.consume() // =
			if t.peek() != nil && t.peek().token_type != Int {
				panic("Expected literal value")
			}
			rhs := Node{node_type: NodeIntLiteral, String: t.consume().value} // 69
			stmts.append(&Node{node_type: NodeLet, lhs: &lhs, rhs: &rhs})

		default:
			panic("Can't parse token " + string(t.peek().token_type) + " " + t.peek().value)
		}
	}

	return &stmts
}
