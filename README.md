## Overview

A simple language written in Go to fulfil both my current interest in doing things closer to the CPU (i.e., writing more asssembly), and learning Go. The motivation will explain two things, 1; why the go code is of questionable quality and 2; why I will probably never finish it.

Feel free to do whatever you like with it.

Inspired by [https://github.com/orosmatthew/hydrogen-cpp]()

```
let x = 5 + 100 / 2 - 2 * (3 + 3)  
exit x
```

I've never done anything with EBNF before but here's an attempt to describe the grammer. I'll try to keep it up to date as I don't actually use it to generate a parser (given this is purely a learning exercise!)

``` ebnf
program
  : statement+
  ;

scope
  : '{' statement+ '}'
  ;

term
  : integer
  | identifier
  | paren_expr
  | function
  ;

test
  : term '<' term
  | term '==' term
  | term '>' term
  ;

expr
  : term '+' term
  | term '-' term
  | term '*' term
  | term '/' term
  ;

paren_expr
  : '(' expr ')'
  ;

statement
  : 'exit' [expr]
  | 'let' identifier '=' expr
  | identifier ':=' expr
  | scope
  | 'if' test scope
  | 'for' test scope
  | function
  ;

```
