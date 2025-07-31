package main

import (
	"assembler/code"
	"assembler/parser"
	"fmt"
	"os"
	"strconv"
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

	//first pass to build symbol table of labels
	pc := 0
	for p.HasMoreCommands() {
		p.Advance()
		switch p.CommandType() {
		case parser.A_COMMAND, parser.C_COMMAND:
			pc++
		case parser.L_COMMAND:
			symbols[p.Symbol()] = pc
		}
	}

	// second pass to assemble, reinitialise parser
	_, err = inf.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	p = parser.NewParser(inf)

	nextRamAddr := symbols["R15"] + 1

	for p.HasMoreCommands() {
		p.Advance()
		switch p.CommandType() {
		case parser.A_COMMAND:
			s := p.Symbol()

			constant, err := strconv.Atoi(s)
			if err == nil {
				outf.WriteString(asAInstruction(constant))
			} else if value, found := symbols[s]; found {
				outf.WriteString(asAInstruction(value))
			} else {
				outf.WriteString(asAInstruction(nextRamAddr))
				symbols[s] = nextRamAddr
				nextRamAddr++
			}

		case parser.C_COMMAND:
			outf.WriteString(fmt.Sprintf("111%07b%03b%03b\n", code.Comp(p.Comp()), code.Dest(p.Dest()), code.Jump(p.Jump())))
		}
	}
}

func asAInstruction(i int) string {
	return fmt.Sprintf("0%015b\n", i)
}

var symbols = map[string]int{
	"SP":     0x0000,
	"LCL":    0x0001,
	"ARG":    0x0002,
	"THIS":   0x0003,
	"THAT":   0x0004,
	"R0":     0x0000,
	"R1":     0x0001,
	"R2":     0x0002,
	"R3":     0x0003,
	"R4":     0x0004,
	"R5":     0x0005,
	"R6":     0x0006,
	"R7":     0x0007,
	"R8":     0x0008,
	"R9":     0x0009,
	"R10":    0x000A,
	"R11":    0x000B,
	"R12":    0x000C,
	"R13":    0x000D,
	"R14":    0x000E,
	"R15":    0x000F,
	"SCREEN": 0x4000,
	"KBD":    0x6000,
}
