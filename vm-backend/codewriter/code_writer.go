package codewriter

import (
	"fmt"
	"io"
	"log"
)

type CodeWriter struct {
	w io.WriteCloser

	// used to scope static variables
	currentFilename string
	// used to make external labels, scoped to the function they are defined in, globally unique
	currentFuncName string
	// used to ensure internal labels are globally unique
	labelCount int
}

func NewCodeWriter(w io.WriteCloser) CodeWriter {
	return CodeWriter{w: w}
}

func (c *CodeWriter) SetFilename(filename string) {
	c.currentFilename = filename
}

func (c *CodeWriter) WriteInit() {
	c.writeLines(
		"@256",
		"D=A",
		"@SP",
		"M=D",
	)
	c.SetFilename("Sys.vm")
	c.WriteCall("Sys.init", 0)
}

func (c *CodeWriter) WriteCall(funcName string, nArgs int) {
	retAddr := c.uniqueLabel()

	// push return address to stack
	c.writeLines(
		fmt.Sprintf("@%v", retAddr),
		"D=A",
	)
	c.pushFromDToStack()

	c.pushLabelMemToStack("LCL")
	c.pushLabelMemToStack("ARG")
	c.pushLabelMemToStack("THIS")
	c.pushLabelMemToStack("THAT")

	// now we have this stuff on the stack above the called func arguments,
	// so let's set ARG to be SP minus all the above (taking us to the end of the func arg block)
	// minus the size of the arg block (nArgs)
	// i.e. set arg to SP-5-nArgs
	c.writeLines(
		"@SP",
		"D=M",
		fmt.Sprintf("@%v", 5+nArgs),
		"D=D-A",
		"@ARG",
		"M=D",
	)

	// set LCL = SP
	// when we goto f, the first thing that will be run is LCL[i]=0
	c.writeLines(
		"@SP",
		"D=M",
		"@LCL",
		"M=D",
	)

	// goto f and label return
	c.writeLines(
		fmt.Sprintf("@%v", funcName),
		"0;JMP",
		fmt.Sprintf("(%v)", retAddr),
	)
}

func (c *CodeWriter) pushLabelMemToStack(l string) {
	c.writeLines(
		fmt.Sprintf("@%v", l),
		"D=M",
	)
	c.pushFromDToStack()
}

func (c *CodeWriter) WriteReturn() {
	c.writeLines(
		"@LCL", // ret address is 5 before lcl
		"D=M",
		"@5",
		"A=D-A",
		"D=M",
		"@R14", // save RET to R14
		"M=D",

		// pop a value from the stack and save to ARG, which will become the return
		// value when control is returned to the caller
		"@SP",
		"AM=M-1", // strictly this assignment to M isn't necessary as we're about to nuke SP
		"D=M",
		"@ARG", // get ARG
		"A=M",  // get what it points to
		"M=D",  // set the value of what it points to to the return value
		"@ARG", // get ARG again
		"D=M",  // get what it points to
		"@SP",
		"M=D+1", // set the caller's stack pointer to one after the return location

		//restore that, this, arg, lcl
		"@LCL",
		"A=M-1", // that pointer
		"D=M",   // get the value of that
		"@THAT",
		"M=D",

		"@LCL", // local **
		"D=M",  // local *
		"@2",
		"A=D-A", // THIS *
		"D=M",   // THIS val
		"@THIS",
		"M=D",

		"@LCL",
		"D=M",
		"@3",
		"A=D-A",
		"D=M",
		"@ARG",
		"M=D",

		"@LCL",
		"D=M",
		"@4",
		"A=D-A",
		"D=M",
		"@LCL",
		"M=D",

		// goto ret
		"@R14",
		"A=M",
		"0;JMP",
	)
}

func (c *CodeWriter) WriteFunction(funcName string, nLocals int) {
	c.currentFuncName = funcName
	c.writeLines(
		fmt.Sprintf("(%v)", c.currentFuncName),
	)
	for range nLocals {
		c.writePush("constant", 0)
	}
}

func (c *CodeWriter) WriteLabel(l string) {
	c.writeLines(fmt.Sprintf("(%v$%v)", c.currentFuncName, l))
}
func (c *CodeWriter) WriteGoto(l string) {
	c.writeLines(
		fmt.Sprintf("@%v$%v", c.currentFuncName, l),
		"0;JMP",
	)
}

// if top of stack is not zero, goto
func (c *CodeWriter) WriteIfGoto(l string) {
	c.writeLines(
		"@SP",
		"AM=M-1",
		"D=M",
		fmt.Sprintf("@%v$%v", c.currentFuncName, l),
		"D;JNE",
	)
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
		c.writePopToBaseIndex("@LCL", index)
	case "argument":
		c.writePopToBaseIndex("@ARG", index)
	case "this":
		c.writePopToBaseIndex("@THIS", index)
	case "that":
		c.writePopToBaseIndex("@THAT", index)
	case "temp":
		c.writePopToDirect(fmt.Sprintf("@R%v", 5+index))
	case "pointer":
		c.writePopToDirect(fmt.Sprintf("@R%v", 3+index))
	case "static":
		c.writePopToDirect(fmt.Sprintf("@%v.%v", c.currentFilename, index))
	default:
		log.Fatal("pop unimplemented for segment: " + segment)
	}
}

func (c *CodeWriter) writePopToDirect(atSegment string) {
	c.writeLines(
		"@SP",
		"AM=M-1",
		"D=M", // value on top of stack in D
		atSegment,
		"M=D",
	)
}

func (c *CodeWriter) writePopToBaseIndex(atSegment string, index int) {
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
	switch segment {
	case "constant":
		c.writeLines(fmt.Sprintf("@%v", index), "D=A")
	case "local":
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
	case "pointer":
		c.writeLines(
			fmt.Sprintf("@R%v", 3+index),
			"D=M",
		)
	case "static":
		c.writeLines(
			fmt.Sprintf("@%v.%v", c.currentFilename, index),
			"D=M",
		)
	default:
		log.Fatal("push unimplemented for segment: " + segment)
	}

	c.pushFromDToStack()
}

func (c *CodeWriter) pushFromDToStack() {
	c.writeLines("@SP", "A=M", "M=D", "@SP", "M=M+1")
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
	label := fmt.Sprintf("label%d", c.labelCount)
	c.labelCount++
	return label
}
