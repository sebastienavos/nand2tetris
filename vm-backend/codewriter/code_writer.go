package codewriter

import (
	"fmt"
	"io"
	"log"
)

type CodeWriter struct {
	w               io.WriteCloser
	currentFilename string
	labelCount      int
}

func NewCodeWriter(w io.WriteCloser) CodeWriter {
	return CodeWriter{w: w}
}

func (c CodeWriter) SetFilename(filename string) {
	c.currentFilename = filename
}

func (c CodeWriter) WriteArithmetic(cmd string) {
	switch cmd {
	case "add":
		c.writeBinaryOperator("D=D+M")
	case "sub":
		c.writeBinaryOperator("D=M-D")
	case "and":
		c.writeBinaryOperator("D=D&M")
	case "or":
		c.writeBinaryOperator("D=D|M")
	case "neg":
		c.writeUnaryOperator("M=-M")
	case "not":
		c.writeUnaryOperator("M=!M")
	case "eq":
		c.writeComp("D;JEQ")
	case "gt":
		c.writeComp("D;JGT")
	case "lt":
		c.writeComp("D;JLT")
	default:
		log.Fatal("unsupported cmd")
	}
}

func (c CodeWriter) writeComp(cmd string) {
	// do x-y
	c.writeLines("@SP", "M=M-1", "A=M", "D=M", "@SP", "M=M-1", "A=M", "D=M-D")
	// define a block to jump to iff comparison is true and effect jump
	t := c.uniqueLabel("true")
	c.writeLines(fmt.Sprintf("@%v", t), cmd)
	// if false, we continue to push 0, but must skip the true block so define another label for the continuation
	cont := c.uniqueLabel("continue")
	c.writeLines("@SP", "A=M", "M=0", fmt.Sprintf("@%v", cont), "0;JMP") // push zero to x then skip pushing -1
	// implement true block
	c.writeLines(fmt.Sprintf("(%v)", t))
	c.writeLines("@SP", "A=M", "M=-1")
	// increment SP if true or false
	c.writeLines("@SP", "M=M+1")
}

func (c CodeWriter) WritePushPop(cmd, segment string, index int) {
	if cmd != "push" {
		log.Fatal("cmd not push")
	}
	if segment != "constant" {
		log.Fatal("segment not constant")
	}

	c.writeLines("@index", "D=A", "@SP", "A=M", "M=D", "@SP", "M=M+1")
}

func (c CodeWriter) Close() { c.w.Close() }

// operate on some memory location M and reassign to M
func (c *CodeWriter) writeUnaryOperator(cmd string) {
	c.writeLines("@SP", "M=M-1", "A=M", cmd, "@SP", "M=M+1")
}

// operate on M (SP-2) and D (SP-1) then assign to D
func (c *CodeWriter) writeBinaryOperator(cmd string) {
	c.writeLines("@SP", "M=M-1", "A=M", "D=M", "@SP", "M=M-1", "A=M", cmd, "M=D", "@SP", "M=M+1")
}

func (c *CodeWriter) writeLines(lines ...string) {
	for _, line := range lines {
		fmt.Fprintln(c.w, line)
	}
}

func (c *CodeWriter) uniqueLabel(base string) string {
	label := fmt.Sprintf("%s.%d", base, c.labelCount)
	c.labelCount++
	return label
}
