package generator

import (
	"fmt"
	"os"
	"strconv"

	"longden.me/blang/parser"
)

type Variable struct {
	name string
	loc  int
}

type Stack []int

type String struct {
	name  string
	value string
}

type Generator struct {
	vars        []Variable
	stack_size  int
	scopes      Stack
	output      string
	label_count int
	strings     []String
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

func (g *Generator) gen_term(node *parser.Node) {
	if node.Type == parser.NodeIntLiteral {
		g.output += "    mov rax, " + node.Value + "\n"
		g.output += g.push("rax", "push literal on stack")
	} else if node.Type == parser.NodeStringLiteral {
		label := g.create_label()
		g.strings = append(g.strings, String{name: label, value: node.Value})
		g.output += g.push(label, "string")

	} else if node.Type == parser.NodeIdentifier {
		variable := g.find_var(node.Value)
		if variable == nil {
			panic("No such variable, '" + node.Value + "'")
		}
		// found variable, get location
		g.output += g.push("qword [rsp + "+fmt.Sprint((g.stack_size-variable.loc)*8)+"]", "push "+variable.name+" on stack")
	} else if node.Type == parser.NodeAdd {
		g.gen_term(node.Rhs)
		g.gen_term(node.Lhs)
		g.output += g.pop("rax")
		g.output += g.pop("rbx")
		g.output += "    add rax, rbx\n"
		g.output += g.push("rax", "+")
	} else if node.Type == parser.NodeSub {
		g.gen_term(node.Rhs)
		g.gen_term(node.Lhs)
		g.output += g.pop("rax")
		g.output += g.pop("rbx")
		g.output += "    sub rax, rbx\n"
		g.output += g.push("rax", "-")
	} else if node.Type == parser.NodeMulti {
		g.gen_term(node.Rhs)
		g.gen_term(node.Lhs)
		g.output += g.pop("rax")
		g.output += g.pop("rbx")
		g.output += "    mul rbx\n"
		g.output += g.push("rax", "*")
	} else if node.Type == parser.NodeDiv {
		g.gen_term(node.Rhs)
		g.gen_term(node.Lhs)
		g.output += "    mov rdx, 0\n"
		g.output += g.pop("rax")
		g.output += g.pop("rbx")
		g.output += "    div rbx\n"
		g.output += g.push("rax", "/")
	} else {
		panic("error parsing expression")
	}
}

func (g *Generator) gen_test(node *parser.Node) string {
	g.gen_term(node.Lhs)
	g.gen_term(node.Rhs)
	g.output += g.pop("rax")
	g.output += g.pop("rbx")
	g.output += "    cmp rbx, rax\n"
	switch node.Type {
	case parser.NodeLt:
		return "jl"
	case parser.NodeGt:
		return "jg"
	case parser.NodeEq:
		return "je"
	}
	return "je"
}

func (g *Generator) gen_inverse_test(node *parser.Node) string {
	g.gen_term(node.Lhs)
	g.gen_term(node.Rhs)
	g.output += g.pop("rax")
	g.output += g.pop("rbx")
	g.output += "    cmp rbx, rax\n"
	switch node.Type {
	case parser.NodeLt:
		return "jge"
	case parser.NodeGt:
		return "jle"
	case parser.NodeEq:
		return "jne"
	}
	return "jle"
}

func (g *Generator) create_label() string {
	label := "label" + strconv.Itoa(g.label_count)
	g.label_count++
	return label
}

func (g *Generator) gen_expr(node *parser.Node) {
	switch node.Type {
	case parser.NodeExit:
		g.gen_term(node.Lhs)
		//g.output += "    mov rax, 0x2000001 ; exit system call\n"
		g.output += "    mov rax, 60 ; exit system call\n"
		g.output += g.pop("rdi")
		g.output += "    syscall\n"
	case parser.NodeLet:
		variable := g.find_var(node.Lhs.Value)
		if variable != nil {
			panic("Variable already declared")
		}
		g.gen_term(node.Rhs) // store value on stack

		g.vars = append(g.vars, Variable{name: node.Lhs.Value, loc: g.stack_size})
	case parser.NodeScope:
		g.gen_scope(node)
	case parser.NodeIf:
		g.output += "    ;if\n"
		test := g.gen_inverse_test(node.Lhs)
		label := g.create_label()
		g.output += "    " + test + " " + label + "\n"
		g.gen_scope(node)
		g.output += "    ;endif\n" + label + ":\n"
	case parser.NodeAssign:
		g.output += "    ; assignment\n"
		variable := g.find_var(node.Lhs.Value)
		if variable == nil {
			panic("Attempted assignment to undeclared variable")
		}
		g.gen_term(node.Rhs)
		g.output += g.pop("rax")
		g.output += "    mov qword [rsp + " + fmt.Sprint((g.stack_size-variable.loc)*8) + "], rax\n"

	case parser.NodeFor:
		g.output += "    ;for\n"
		label_start := g.create_label()
		label_end := g.create_label()
		// test if we should enter loop
		test := g.gen_inverse_test(node.Lhs)
		g.output += "    " + test + " " + label_end + "\n"
		g.output += label_start + ":\n"
		g.gen_scope(node)
		test = g.gen_test(node.Lhs)
		g.output += "    " + test + " " + label_start + "\n"
		g.output += "    ; endfor\n" + label_end + ":\n"
	case parser.NodePrint:
		g.gen_term(node.Lhs)
		g.output += g.pop("rsi")
		g.output += "    xor rdx, rdx\n"
		g.output += "    mov rbp, rsi\n"
		label := g.create_label()
		label_end := g.create_label()
		g.output += label + ":\n"
		g.output += "    cmp [rbp], byte 0\n"
		g.output += "    jz " + label_end + "\n"
		g.output += "    inc rbp\n"
		g.output += "    inc rdx\n"
		g.output += "    jmp " + label + "\n"
		g.output += label_end + ":\n"
		g.output += "    mov rax, 1 ; sys_write\n"
		g.output += "    mov rdi, 1 ; stdout\n"
		g.output += "    syscall\n"

	default:
		panic("Can't generate expression")
	}
}

func (g *Generator) gen_scope(node *parser.Node) {
	g.begin_scope()
	for i := 0; i < len(node.Stmts.Statements); i++ {
		g.gen_expr(&node.Stmts.Statements[i])
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
	if len(g.vars) > 0 {
		g.vars = g.vars[:len(g.vars)-pop_count]
	}
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

func (g *Generator) assemble(stmts *parser.StatementSequence) {
	// g.output = "global _main\nsection .text\n_main:\n"
	g.output = "global _start\nsection .text\n_start:\n"

	for i := 0; i < len(stmts.Statements); i++ {
		g.gen_expr(&stmts.Statements[i])
	}

	// g.output += "    mov rax, 0x2000001 ; exit system call\n"
	g.output += "    mov rax, 60 ; exit system call\n"
	g.output += "    mov rdi, 0\n"
	g.output += "    syscall\n"

	g.output += "section .data\n"
	for i := 0; i < len(g.strings); i++ {
		g.output += g.strings[i].name + " db \"" + g.strings[i].value + "\", 0\n"
	}
}

func Generate(stmts *parser.StatementSequence, fn string) {
	g := Generator{}
	g.assemble(stmts)

	if err := os.WriteFile(fn, []byte(g.output), 0644); err != nil {
		fmt.Println(err)
	}
}
