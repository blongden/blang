package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"longden.me/blang/generator"
	"longden.me/blang/parser"
	"longden.me/blang/tokeniser"
)

type VarType int

const (
	Int VarType = iota
	String
)

type Variable struct {
	Name string
	Type VarType
}

type TypeChecker struct {
	variables []Variable
}

func (tc *TypeChecker) GetType(node *parser.Node) (*VarType, error) {
	ty := Int
	if node.Type == parser.NodeStringLiteral {
		ty = String
	} else if node.Type == parser.NodeIdentifier {
		for _, v := range tc.variables {
			if v.Name == node.Value {
				ty = v.Type
				return &ty, nil
			}
		}

		return nil, fmt.Errorf("compiler error, variable not in scope")
	} else if node.Type == parser.NodeAdd {
		lhs, err := tc.GetType(node.Lhs)
		if err != nil {
			return nil, err
		}
		rhs, err := tc.GetType(node.Rhs)
		if err != nil {
			return nil, err
		}

		if lhs != rhs {
			return nil, fmt.Errorf("can't add variables of differing types")
		}
	}

	return &ty, nil
}

func (tc *TypeChecker) CheckNode(node *parser.Node) (*parser.Node, error) {
	if node == nil {
		return nil, nil
	}

	if node.Stmts != nil {
		scope := TypeChecker{variables: tc.variables}
		scope.TypeCheck(node.Stmts)
	}

	rhs, _ := tc.CheckNode(node.Rhs)
	lhs, _ := tc.CheckNode(node.Lhs)
	if node.Type == parser.NodeLet {
		// infer the type
		ty, err := tc.GetType(rhs)
		if err != nil {
			return nil, err
		}
		tc.variables = append(tc.variables, Variable{Name: lhs.Value, Type: *ty})
	}

	if node.Type == parser.NodeAssign {
		lhs, err := tc.GetType(node.Lhs)
		if err != nil {
			return nil, err
		}
		rhs, err := tc.GetType(node.Rhs)
		if err != nil {
			return nil, err
		}

		if lhs != rhs {
			return nil, fmt.Errorf("mismatched type when attempting to reassign variable")
		}
	}
	return node, nil
}

func (tc *TypeChecker) TypeCheck(seq *parser.StatementSequence) error {
	for i := 0; i < len(seq.Statements); i++ {
		_, err := tc.CheckNode(&seq.Statements[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	output := flag.String("o", "out", "output file name")
	flag.Parse()
	source := flag.Arg(0)
	if source == "" {
		fmt.Println("No source file specified!")
		os.Exit(1)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	tokens, err := tokeniser.Tokenise(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Token error: %s\n", err)
		os.Exit(3)
	}
	p := parser.Parser{Tokens: tokens}
	ast, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %s\n", err)
		os.Exit(4)
	}
	tc := TypeChecker{}
	err = tc.TypeCheck(ast)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Type error: %s\n", err)
		os.Exit(5)
	}

	asm_fn := *output + ".a"
	o_fn := *output + ".o"
	generator.Generate(ast, asm_fn)

	// cmd := exec.Command("nasm", "-f", "macho64", "test.a", "-o", "test.o")
	cmd := exec.Command("nasm", "-f", "elf64", asm_fn, "-o", o_fn)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}

	// cmd = exec.Command("ld", "-macosx_version_min", "13.5.0", "-L/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/lib", "-lSystem", "-o", "test", "test.o")
	cmd = exec.Command("ld", "-o", *output, o_fn)
	cmd.Stderr = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println(cmd)
		fmt.Println(err)
	}
	cmd.Run()
}
