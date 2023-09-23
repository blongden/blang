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
	If
	Println
)

type Token struct {
	token_type TokenType
	value      string
}

type Source struct {
	src []byte
	sp  int
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
	return b
}

func tokenise(data []byte) Tokens {
	tokens := []Token{}

	src := Source{src: data}

	buf := ""
	for src.peek() != 0 {
		if unicode.IsLetter(rune(src.peek())) {
			for unicode.IsLetter(rune(src.peek())) || unicode.IsNumber(rune(src.peek())) {
				buf += string(src.consume())
			}
			switch buf {
			case "exit":
				tokens = append(tokens, Token{token_type: Exit})
			case "let":
				tokens = append(tokens, Token{token_type: Let})
			case "if":
				tokens = append(tokens, Token{token_type: If})
			case "println":
				tokens = append(tokens, Token{token_type: Println})
			default:
				tokens = append(tokens, Token{token_type: Identifier, value: buf})
			}
		} else if unicode.IsDigit(rune(src.peek())) {
			for unicode.IsDigit(rune(src.peek())) {
				buf += string(src.consume())
			}

			tokens = append(tokens, Token{token_type: Int, value: buf})
		} else if unicode.IsSpace(rune(src.peek())) {
			src.consume()
		} else if string(src.peek()) == "=" {
			src.consume()
			if string(src.peek()) == "=" {
				src.consume()
				tokens = append(tokens, Token{token_type: Eq})
			} else {
				tokens = append(tokens, Token{token_type: Assign})
			}
		} else if string(src.peek()) == "+" {
			src.consume()
			tokens = append(tokens, Token{token_type: Plus})
		} else if string(src.peek()) == "-" {
			src.consume()
			tokens = append(tokens, Token{token_type: Minus})
		} else if string(src.peek()) == "*" {
			src.consume()
			tokens = append(tokens, Token{token_type: Star})
		} else if string(src.peek()) == "/" {
			src.consume()
			tokens = append(tokens, Token{token_type: Fslash})
		} else if string(src.peek()) == "(" {
			src.consume()
			tokens = append(tokens, Token{token_type: Lparen})
		} else if string(src.peek()) == ")" {
			src.consume()
			tokens = append(tokens, Token{token_type: Rparen})
		} else if string(src.peek()) == "{" {
			src.consume()
			tokens = append(tokens, Token{token_type: Lcurly})
		} else if string(src.peek()) == "}" {
			src.consume()
			tokens = append(tokens, Token{token_type: Rcurly})
		} else {
			panic(fmt.Sprintf("No idea what this is yet at position %d (%c)", src.sp, src.src[src.sp]))
		}
		buf = ""
	}

	return Tokens{tokens: tokens}
}
