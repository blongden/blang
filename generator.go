package main

import (
	"fmt"
	"os"
)

type Variable struct {
	name string
	loc  int
}

type Stack []int

type Generator struct {
	vars        []Variable
	stack_size  int
	scopes      Stack
	output      string
	label_count int
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
			panic("No such variable, '" + node.value + "'")
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
	case NodeScope:
		g.gen_scope(node)
	case NodeIf:
		g.output += "    ;if\n"
		g.gen_term(node.lhs)
		g.output += g.pop("rax")
		g.output += "    test rax, rax\n"

		label := "label" + fmt.Sprint(g.label_count)
		g.label_count++

		g.output += "    jz " + label + "\n"
		g.gen_scope(node)
		g.output += "    ;endif\n" + label + ":\n"

	default:
		panic("Can't generate expression")
	}
}

func (g *Generator) gen_scope(node *Node) {
	g.begin_scope()
	for i := 0; i < len(node.stmts.statements); i++ {
		g.gen_expr(&node.stmts.statements[i])
	}
	g.end_scope()
}

func (g *Generator) begin_scope() {
	g.output += "    ; scope begins\n"
	g.scopes = append(g.scopes, g.stack_size)
}

func (g *Generator) end_scope() {
	target_size := g.scopes[len(g.scopes)-1]
	pop_count := len(g.vars) - target_size
	g.output += "    ; scope ends\n"
	g.output += "    add rsp, " + fmt.Sprint(pop_count*8) + "\n"
	g.stack_size -= pop_count
	g.vars = g.vars[:len(g.vars)-pop_count]
	g.scopes = g.scopes[:len(g.scopes)-1]
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
