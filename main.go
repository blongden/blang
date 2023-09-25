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

func main() {
	flag.Parse()
	source := flag.Args()
	if len(source) == 0 {
		fmt.Println(flag.ErrHelp)
		os.Exit(1)
	}
	data, err := os.ReadFile(source[0])
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
	generator.Generate(ast)

	cmd := exec.Command("nasm", "-f", "macho64", "test.a", "-o", "test.o")
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}

	cmd = exec.Command("ld", "-macosx_version_min", "13.5.0", "-L/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/lib", "-lSystem", "-o", "test", "test.o")
	cmd.Run()
}
