## Overview

A simple language written in Go to fulfil both my current interest in doing things closer to the CPU (i.e., writing more asssembly), and learning Go. The motivation will explain two things, 1; why the go code is of questionable quality and 2; why I will probably never finish it.

Feel free to do whatever you like with it.

```
let x = 5 + 100 / 2 - 2 * (3 + 3)  
exit x
```

``` ebnf
program
  : statement+
  ;

statement
  : exit [expr]
  | let identifier '=' expr
  ;

paren_expr
  : '(' expr ')'
  ;

term
  : integer
  | identifier
  | paren_expr
  ;

expr
  : term '+' term
  | term '-' term
  | term '*' term
  | term '/' term
  ;

```