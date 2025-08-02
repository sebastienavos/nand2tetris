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

func (c *CodeWriter) SetFilename(filename string) {
	c.currentFilename = filename
}

func (c *CodeWriter) WriteArithmetic(cmd string) {
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

func (c *CodeWriter) writeComp(comparisonJump string) {
	trueLabel := c.uniqueLabel()
	contLabel := c.uniqueLabel()
	c.writeLines(
		// do x-y
		"@SP",
		"AM=M-1",
		"D=M",
		"@SP",
		"AM=M-1",
		"D=M-D",
		// define a block to jump to iff comparison is true and effect jump
		fmt.Sprintf("@%v", trueLabel),
		comparisonJump,
		// if false, we continue to push 0, but must skip the true block so jump to the continuation
		"@SP",
		"A=M",
		"M=0",
		fmt.Sprintf("@%v", contLabel),
		"0;JMP",
		// implement true block
		fmt.Sprintf("(%v)", trueLabel),
		"@SP",
		"A=M",
		"M=-1",
		//end of true block
		fmt.Sprintf("(%v)", contLabel),
		"@SP",
		"M=M+1",
	)
}

func (c *CodeWriter) WritePushPop(cmd, segment string, index int) {
	switch cmd {
	case "push":
		c.writePush(segment, index)
	case "pop":
		c.writePop(segment, index)
	default:
		log.Fatal("command not push or pop")
	}
}

func (c *CodeWriter) writePop(segment string, index int) {
	switch segment {
	case "local":
		c.writePopTo("@LCL", index)
	case "argument":
		c.writePopTo("@ARG", index)
	case "this":
		c.writePopTo("@THIS", index)
	case "that":
		c.writePopTo("@THAT", index)
	case "temp":
		c.writeLines(
			"@SP",
			"AM=M-1",
			"D=M", // value on top of stack in D
			fmt.Sprintf("@R%v", 5+index),
			"M=D",
		)
	default:
		log.Fatal("pop unimplemented for segment: " + segment)
	}
}

func (c *CodeWriter) writePopTo(atSegment string, index int) {
	c.writeLines(
		atSegment,
		"D=M",
		fmt.Sprintf("@%v", index),
		"D=D+A",
		"@R13",
		"M=D", // location in R13
		"@SP",
		"AM=M-1",
		"D=M", // value on top of stack in D
		"@R13",
		"A=M",
		"M=D", // save to lcl
	)
}

func (c *CodeWriter) writePush(segment string, index int) {
	pushFromDToStack := []string{"@SP", "A=M", "M=D", "@SP", "M=M+1"}

	switch segment {
	case "constant":
		c.writeLines(fmt.Sprintf("@%v", index), "D=A")
	case "local":
		// Set D to the value of the memory at location M(LCL)+index
		c.loadToD("@LCL", index)
	case "argument":
		c.loadToD("@ARG", index)
	case "this":
		c.loadToD("@THIS", index)
	case "that":
		c.loadToD("@THAT", index)
	case "temp":
		c.writeLines(
			fmt.Sprintf("@R%v", 5+index),
			"D=M",
		)
	default:
		log.Fatal("push unimplemented for segment: " + segment)
	}

	c.writeLines(pushFromDToStack...)
}

func (c *CodeWriter) loadToD(atSegment string, index int) {
	c.writeLines(
		atSegment,
		"D=M",
		fmt.Sprintf("@%v", index),
		"A=D+A",
		"D=M",
	)
}

func (c *CodeWriter) Close() { c.w.Close() }

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

func (c *CodeWriter) uniqueLabel() string {
	label := fmt.Sprintf("label.%d", c.labelCount)
	c.labelCount++
	return label
}
