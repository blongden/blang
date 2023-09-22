package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	data, err := os.ReadFile("test.bl")
	if err != nil {
		panic(err)
	}

	tokens := Tokens{tokens: tokenise(data)}
	ast := tokens.parse()
	fmt.Println(ast)
	generate(ast)

	cmd := exec.Command("nasm", "-f", "macho64", "test.a", "-o", "test.o")
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}

	cmd = exec.Command("ld", "-macosx_version_min", "13.5.0", "-L/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/lib", "-lSystem", "-o", "test", "test.o")
	cmd.Run()
}
