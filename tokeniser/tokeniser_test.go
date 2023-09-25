package tokeniser

import (
	"strings"
	"testing"
)

func TestExit(t *testing.T) {
	tokens, _ := Tokenise([]byte("exit"))
	if len(tokens) == 0 || tokens[0].Type != Exit {
		t.Errorf("exit does not generate exit token")
	}
}

func TestValidTokens(t *testing.T) {
	tokens := "1 a abc + - * / < > let exit if for == ( ) { }"
	tokenised, _ := Tokenise([]byte(tokens))
	if len(tokenised) != len(strings.Split(tokens, " ")) {
		t.Errorf("parsed tokens does not match expected total")
	}
}

func TestNewLineIncreasesLineCount(t *testing.T) {
	tokens, _ := Tokenise([]byte("a\nb\nc"))
	if tokens[0].Line != 1 {
		t.Errorf("first token is not on line 1")
	}
	if tokens[1].Line != 2 {
		t.Errorf("first token is not on line 2")
	}
	if tokens[2].Line != 3 {
		t.Errorf("first token is not on line 3")
	}
}

func TestColumnPreserved(t *testing.T) {
	tokens, _ := Tokenise([]byte("let a = b"))
	if len(tokens) != 4 {
		t.Errorf("insufficient tokens generated: %d", len(tokens))
	}

	if tokens[0].Col != 0 {
		t.Errorf("expected column to equal 0 (got %d)", tokens[0].Col)
	}

	if tokens[1].Col != 4 {
		t.Errorf("expected column to equal 4 (got %d)", tokens[1].Col)
	}

	if tokens[2].Col != 6 {
		t.Errorf("expected column to equal 6 (got %d)", tokens[2].Col)
	}

	if tokens[3].Col != 8 {
		t.Errorf("expected column to equal 8 (got %d)", tokens[3].Col)
	}
}

func TestUnrecognisedTokenGeneratesError(t *testing.T) {
	_, err := Tokenise([]byte("§"))
	if err == nil {
		t.Errorf("expected error from tokeniser")
	}
}
