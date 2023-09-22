package main

import (
	"fmt"
	"os"
)

type Variable struct {
	name string
	loc  int
}

type Generator struct {
	vars       []Variable
	stack_size int
	output     string
}

func (g *Generator) find_var(s string) *Variable {
	var variable *Variable
	for _, v := range g.vars {
		if v.name == s {
			variable = &v
			break
		}
	}
	return variable
}

func (g *Generator) gen_term(node *Node) {
	if node.node_type == NodeIntLiteral {
		g.output += "    mov rax, " + node.value + "\n"
		g.output += g.push("rax", "push literal on stack")
	} else if node.node_type == NodeIdentifier {
		variable := g.find_var(node.value)
		if variable == nil {
			panic("No such variable, '" + node.lhs.value + "'")
		}
		// found variable, get location
		g.output += g.push("qword [rsp + "+fmt.Sprint((g.stack_size-variable.loc)*8)+"]", "push "+variable.name+" on stack")
	} else if node.node_type == NodeAdd {
		g.gen_term(node.rhs)
		g.gen_term(node.lhs)
		g.output += g.pop("rax")
		g.output += g.pop("rbx")
		g.output += "    add rax, rbx\n"
		g.output += g.push("rax", "+")
	} else if node.node_type == NodeSub {
		g.gen_term(node.rhs)
		g.gen_term(node.lhs)
		g.output += g.pop("rax")
		g.output += g.pop("rbx")
		g.output += "    sub rax, rbx\n"
		g.output += g.push("rax", "-")
	} else if node.node_type == NodeMulti {
		g.gen_term(node.rhs)
		g.gen_term(node.lhs)
		g.output += g.pop("rax")
		g.output += g.pop("rbx")
		g.output += "    mul rbx\n"
		g.output += g.push("rax", "*")
	} else if node.node_type == NodeDiv {
		g.gen_term(node.rhs)
		g.gen_term(node.lhs)
		g.output += "    mov rdx, 0\n"
		g.output += g.pop("rax")
		g.output += g.pop("rbx")
		g.output += "    div rbx\n"
		g.output += g.push("rax", "/")
	} else {
		panic("error parsing expression")
	}
}

func (g *Generator) gen_expr(node *Node) {
	switch node.node_type {
	case NodeExit:
		g.gen_term(node.lhs)
		g.output += "    mov rax, 0x2000001 ; exit system call\n"
		g.output += g.pop("rdi")
		g.output += "    syscall\n"
	case NodeLet:
		variable := g.find_var(node.lhs.value)
		if variable != nil {
			panic("Variable already declared")
		}
		g.gen_term(node.rhs) // store value on stack
		g.vars = append(g.vars, Variable{name: node.lhs.value, loc: g.stack_size})
	default:
		panic("Can't parse expression")
	}
}

func (g *Generator) push(reg string, comment string) string {
	output := "    push " + reg + " ; " + comment + "\n"
	g.stack_size++
	return output
}

func (g *Generator) pop(reg string) string {
	output := "    pop " + reg + "\n"
	g.stack_size--
	return output
}

func (g *Generator) assemble(stmts *StatementSequence) {
	g.output = "global _main\n_main:\n"

	for i := 0; i < len(stmts.statements); i++ {
		g.gen_expr(&stmts.statements[i])
	}

	g.output += "    mov rax, 0x2000001 ; exit system call\n"
	g.output += "    mov rdi, 0\n"
	g.output += "    syscall\n"
}

func generate(stmts *StatementSequence) {
	g := Generator{}
	g.assemble(stmts)

	if err := os.WriteFile("test.a", []byte(g.output), 0644); err != nil {
		fmt.Println(err)
	}
}
