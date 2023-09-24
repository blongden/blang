package main

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
	token_type TokenType
	value      string
	line       int
	col        int
}

type Source struct {
	src    []byte
	tokens []Token
	sp     int
	line   int
	col    int
}

func (s *Source) peek() byte {
	if s.sp >= len(s.src) {
		return 0
	}
	return s.src[s.sp]
}

func (s *Source) consume() byte {
	b := s.src[s.sp]
	s.sp++
	s.col++
	return b
}

func (s *Source) append(token Token) {
	token.col = s.col
	token.line = s.line
	s.tokens = append(s.tokens, token)
}

func tokenise(data []byte) []Token {
	src := Source{src: data}

	buf := ""
	for src.peek() != 0 {
		if unicode.IsLetter(rune(src.peek())) {
			for unicode.IsLetter(rune(src.peek())) || unicode.IsNumber(rune(src.peek())) {
				buf += string(src.consume())
			}
			switch buf {
			case "exit":
				src.append(Token{token_type: Exit})
			case "let":
				src.append(Token{token_type: Let})
			case "if":
				src.append(Token{token_type: If})
			case "for":
				src.append(Token{token_type: For})
			default:
				src.append(Token{token_type: Identifier, value: buf})
			}
		} else if unicode.IsDigit(rune(src.peek())) {
			for unicode.IsDigit(rune(src.peek())) {
				buf += string(src.consume())
			}

			src.append(Token{token_type: Int, value: buf})
		} else if unicode.IsSpace(rune(src.peek())) {
			if src.peek() == 20 {
				src.line++
				src.col = 0
			}
			src.consume()
		} else if string(src.peek()) == "=" {
			src.consume()
			if string(src.peek()) == "=" {
				src.consume()
				src.append(Token{token_type: Eq})
			} else {
				src.append(Token{token_type: Assign})
			}
		} else if string(src.peek()) == "+" {
			src.consume()
			src.append(Token{token_type: Plus})
		} else if string(src.peek()) == "-" {
			src.consume()
			src.append(Token{token_type: Minus})
		} else if string(src.peek()) == "*" {
			src.consume()
			src.append(Token{token_type: Star})
		} else if string(src.peek()) == "/" {
			src.consume()
			src.append(Token{token_type: Fslash})
		} else if string(src.peek()) == "(" {
			src.consume()
			src.append(Token{token_type: Lparen})
		} else if string(src.peek()) == ")" {
			src.consume()
			src.append(Token{token_type: Rparen})
		} else if string(src.peek()) == "{" {
			src.consume()
			src.append(Token{token_type: Lcurly})
		} else if string(src.peek()) == "}" {
			src.consume()
			src.append(Token{token_type: Rcurly})
		} else if string(src.peek()) == "<" {
			src.consume()
			src.append(Token{token_type: Lt})
		} else if string(src.peek()) == ">" {
			src.consume()
			src.append(Token{token_type: Gt})
		} else {
			panic(fmt.Sprintf("No idea what this is yet at position %d (%c)", src.sp, src.src[src.sp]))
		}
		buf = ""
	}

	return src.tokens
}
