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

func (tc *TypeChecker) GetType(node *parser.Node) VarType {
	ty := Int
	if node.Type == parser.NodeStringLiteral {
		ty = String
	} else if node.Type == parser.NodeIdentifier {
		for _, v := range tc.variables {

			if v.Name == node.Value {
				return v.Type
			}
		}

		panic("Compiler error, variable not in scope")
	} else if node.Type == parser.NodeAdd {
		ty = tc.GetType(node.Lhs)
		if ty == tc.GetType(node.Rhs) {
			return ty
		} else {
			panic("Type mismatch")
		}
	}

	return ty
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
		tc.variables = append(tc.variables, Variable{Name: lhs.Value, Type: tc.GetType(rhs)})
	}

	if node.Type == parser.NodeAssign {
		// does variable exist?
		if tc.GetType(node.Lhs) != tc.GetType(node.Rhs) {
			panic("djshdjshs")
		}
	}
	return node, nil
}

func (tc *TypeChecker) TypeCheck(seq *parser.StatementSequence) error {
	for i := 0; i < len(seq.Statements); i++ {
		tc.CheckNode(&seq.Statements[i])
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
		panic(err)
	}
	tokens, err := tokeniser.Tokenise(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
	p := parser.Parser{Tokens: tokens}
	ast := p.Parse()
	tc := TypeChecker{}
	tc.TypeCheck(ast)

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
