package main

import (
	"fmt"
	"os"
)

func (g *Generator) exit_syscall(code int64) string {
	output := "    mov rax, 0x2000001 ; exit system call\n"
	output += g.pop("rdi")
	output += "    syscall\n"

	return output
}

type Variable struct {
	name string
	loc  int
}

type Generator struct {
	vars       []Variable
	stack_size int
}

func (g *Generator) gen_expr(node *Node) string {
	output := ""
	switch node.node_type {
	case NodeExit:
		if node.lhs.node_type == NodeIntLiteral {
			output += "    mov rax, " + node.lhs.String + "\n"
			output += g.push("rax")
		} else if node.lhs.node_type == NodeIdentifier {
			fmt.Println(node.lhs.String)
			var variable *Variable
			for _, v := range g.vars {
				if v.name == node.lhs.String {
					variable = &v
					break
				}
			}
			if variable == nil {
				panic("No such variable, '" + node.lhs.String + "'")
			}
			// found variable, get location
			fmt.Println("stack size " + fmt.Sprint(g.stack_size))
			fmt.Println("var found at location " + fmt.Sprint(variable.loc))
			fmt.Println("var name " + variable.name)
			output += g.push("qword [rsp + " + fmt.Sprint((g.stack_size-variable.loc)*8) + "]")
		}
	case NodeLet:
		output += "    mov rax, " + node.rhs.String + "\n"
		output += g.push("rax")
		g.vars = append(g.vars, Variable{name: node.lhs.String, loc: g.stack_size})
	}
	return output
}

func (g *Generator) push(reg string) string {
	output := "    push " + reg + "\n"
	g.stack_size++
	return output
}

func (g *Generator) pop(reg string) string {
	output := "    pop " + reg + "\n"
	g.stack_size--
	return output
}

func generate(stmts *StatementSequence) {
	output := "global _main\n_main:\n"
	gen := Generator{}

	for i := 0; i < len(stmts.statements); i++ {
		output += gen.gen_expr(&stmts.statements[i])
	}

	output += gen.exit_syscall(0)

	if err := os.WriteFile("test.a", []byte(output), 0644); err != nil {
		fmt.Println(err)
	}
}
