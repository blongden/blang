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
	String
	Print
	Println
	LetOp
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
	s.tokens = append(s.tokens, token)
}

func Tokenise(data []byte) ([]Token, error) {
	src := source{src: data, line: 1}

	for src.peek() != 0 {
		buf := ""
		t := Token{Col: src.col, Line: src.line}
		if unicode.IsLetter(rune(src.peek())) {
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
			case "print":
				t.Type = Print
			case "println":
				t.Type = Println
			default:
				t.Type = Identifier
				t.Value = buf
			}
		} else if unicode.IsDigit(rune(src.peek())) {
			for unicode.IsDigit(rune(src.peek())) {
				buf += string(src.consume())
			}

			t.Type = Int
			t.Value = buf
		} else if unicode.IsSpace(rune(src.peek())) {
			space := src.consume()
			if space == 10 { // newline
				src.line++
				src.col = 0
			}
			continue
		} else if string(src.peek()) == "\"" {
			src.consume() // string
			for string(src.peek()) != "\"" {
				buf += string(src.consume())
			}
			src.consume()
			t.Type = String
			t.Value = buf
		} else if string(src.peek()) == "=" {
			src.consume()
			if string(src.peek()) == "=" {
				src.consume()
				t.Type = Eq
			} else {
				t.Type = Assign
			}
		} else if string(src.peek()) == "+" {
			src.consume()
			t.Type = Plus
		} else if string(src.peek()) == "-" {
			src.consume()
			t.Type = Minus
		} else if string(src.peek()) == "*" {
			src.consume()
			t.Type = Star
		} else if string(src.peek()) == "/" {
			src.consume()
			t.Type = Fslash

			if string(src.peek()) == "/" {
				// comments like this one, just skip over them
				for src.peek() != 10 && src.peek() != 0 {
					src.consume()
				}
				continue
			}
		} else if string(src.peek()) == "(" {
			src.consume()
			t.Type = Lparen
		} else if string(src.peek()) == ")" {
			src.consume()
			t.Type = Rparen
		} else if string(src.peek()) == "{" {
			src.consume()
			t.Type = Lcurly
		} else if string(src.peek()) == "}" {
			src.consume()
			t.Type = Rcurly
		} else if string(src.peek()) == "<" {
			src.consume()
			t.Type = Lt
		} else if string(src.peek()) == ">" {
			src.consume()
			t.Type = Gt
		} else if string(src.peek()) == ":" {
			src.consume()
			if string(src.peek()) == "=" {
				src.consume()
				t.Type = LetOp
			} else {
				return nil, fmt.Errorf("expected '=' at line %d column %d", src.line, src.col)
			}
		} else {
			return nil, fmt.Errorf("no idea what this is yet at position %d (%c)", src.sp, src.src[src.sp])
		}
		src.append(t)
	}

	return src.tokens, nil
}
