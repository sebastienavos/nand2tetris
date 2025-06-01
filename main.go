package main

import (
	"assembler/code"
	"assembler/parser"
	"fmt"
	"os"
)

func main() {
	in := os.Args[1]
	out := os.Args[2]
	outf, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer outf.Close()

	inf, err := os.Open(in)
	if err != nil {
		panic(err)
	}
	defer inf.Close()

	p := parser.NewParser(inf)

	for p.HasMoreCommands() {
		p.Advance()
		switch p.CommandType() {
		case parser.A_COMMAND:
			outf.WriteString(fmt.Sprintf("0%v\n", p.Symbol()))
		case parser.L_COMMAND:
			panic("unimplemented symbol table")
		case parser.C_COMMAND:
			outf.WriteString(fmt.Sprintf("111%07b%03b%03b\n", code.Comp(p.Comp()), code.Dest(p.Dest()), code.Jump(p.Jump())))
		}
	}
}
