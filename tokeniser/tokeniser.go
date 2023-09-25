package tokeniser

import (
	"fmt"
	"unicode"
)

type TokenType int

const (
	Unused TokenType = iota
	Exit
	Int
	Let
	Assign
	Identifier
	Plus
	Minus
	Star
	Fslash
	Lparen
	Rparen
	Lcurly
	Rcurly
	Eq
	Lt
	Gt
	If
	For
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

type source struct {
	src    []byte
	tokens []Token
	sp     int
	line   int
	col    int
}

func (s *source) peek() byte {
	if s.sp >= len(s.src) {
		return 0
	}
	return s.src[s.sp]
}

func (s *source) consume() byte {
	b := s.src[s.sp]
	s.sp++
	s.col++
	return b
}

func (s *source) append(token Token) {
	if token.Col == 0 {
		token.Col = s.col
	}

	if token.Line == 0 {
		token.Line = s.line
	}

	s.tokens = append(s.tokens, token)
}

func Tokenise(data []byte) []Token {
	src := source{src: data, line: 1}

	buf := ""
	for src.peek() != 0 {
		if unicode.IsLetter(rune(src.peek())) {
			t := Token{Col: src.col, Line: src.line}
			for unicode.IsLetter(rune(src.peek())) || unicode.IsNumber(rune(src.peek())) {
				buf += string(src.consume())
			}
			switch buf {
			case "exit":
				t.Type = Exit
			case "let":
				t.Type = Let
			case "if":
				t.Type = If
			case "for":
				t.Type = For
			default:
				t.Type = Identifier
				t.Value = buf
			}
			src.append(t)
		} else if unicode.IsDigit(rune(src.peek())) {
			for unicode.IsDigit(rune(src.peek())) {
				buf += string(src.consume())
			}

			src.append(Token{Type: Int, Value: buf})
		} else if unicode.IsSpace(rune(src.peek())) {
			if src.peek() == 10 {
				src.line++
				src.col = 0
			}
			src.consume()
		} else if string(src.peek()) == "=" {
			src.consume()
			if string(src.peek()) == "=" {
				src.consume()
				src.append(Token{Type: Eq})
			} else {
				src.append(Token{Type: Assign})
			}
		} else if string(src.peek()) == "+" {
			src.consume()
			src.append(Token{Type: Plus})
		} else if string(src.peek()) == "-" {
			src.consume()
			src.append(Token{Type: Minus})
		} else if string(src.peek()) == "*" {
			src.consume()
			src.append(Token{Type: Star})
		} else if string(src.peek()) == "/" {
			src.consume()
			src.append(Token{Type: Fslash})
		} else if string(src.peek()) == "(" {
			src.consume()
			src.append(Token{Type: Lparen})
		} else if string(src.peek()) == ")" {
			src.consume()
			src.append(Token{Type: Rparen})
		} else if string(src.peek()) == "{" {
			src.consume()
			src.append(Token{Type: Lcurly})
		} else if string(src.peek()) == "}" {
			src.consume()
			src.append(Token{Type: Rcurly})
		} else if string(src.peek()) == "<" {
			src.consume()
			src.append(Token{Type: Lt})
		} else if string(src.peek()) == ">" {
			src.consume()
			src.append(Token{Type: Gt})
		} else {
			panic(fmt.Sprintf("No idea what this is yet at position %d (%c)", src.sp, src.src[src.sp]))
		}
		buf = ""
	}

	return src.tokens
}
