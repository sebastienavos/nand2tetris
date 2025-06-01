package main

import (
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
			outf.WriteString(p.Symbol())
			outf.WriteString("\n")
		case parser.L_COMMAND:
			panic("unimplemented symbol table")
		case parser.C_COMMAND:
			outf.WriteString(fmt.Sprintf("%03b%07b%03b\n", p.Dest(), p.Comp(), p.Jump()))
		}
	}
}
