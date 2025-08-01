package parser

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type CommandType = string

const (
	C_ARITHMETIC CommandType = "C_ARITHMETIC"
	C_PUSH       CommandType = "C_PUSH"
	C_POP        CommandType = "C_POP"
	C_LABEL      CommandType = "C_LABEL"
	C_GOTO       CommandType = "C_GOTO"
	C_IF         CommandType = "C_IF"
	C_FUNCTION   CommandType = "C_FUNCTION"
	C_RETURN     CommandType = "C_RETURN"
	C_CALL       CommandType = "C_CALL"
)

var arithmeticCmds = map[string]struct{}{
	"add": {},
	"sub": {},
	"neg": {},
	"eq":  {},
	"gt":  {},
	"lt":  {},
	"and": {},
	"or":  {},
	"not": {},
}

type parser struct {
	scanner *bufio.Scanner
	// hasMore means next is valid
	hasMore        bool
	currentCommand string
	nextCommand    string
}

func NewParserFromFilename(fileName string) *parser {
	file, err := os.Open(fileName)
	if err != nil {
		panic("could not open file")
	}

	return NewParser(file)
}

func NewParser(reader io.Reader) *parser {
	scanner := bufio.NewScanner(reader)
	parser := parser{scanner: scanner}

	parser.tryFindNextCommand()

	return &parser
}

// Are there more commands in the input?
func (p *parser) HasMoreCommands() bool {
	return p.hasMore
}

// Reads the next command from the input and makes it the current command.
// Should be called only if hasMoreCommands() is true.
// Initially there is no current command.
func (p *parser) Advance() {
	p.currentCommand = p.nextCommand
	p.tryFindNextCommand()
}

func (p *parser) CommandType() CommandType {
	x := strings.Split(p.currentCommand, " ")
	if _, ok := arithmeticCmds[x[0]]; ok {
		return C_ARITHMETIC
	}

	if x[0] == "push" {
		return C_PUSH
	}

	if x[0] == "pop" {
		return C_POP
	}

	panic("")
}

// Returns the first argument of the current command.
// In the case of C_ARITHMETIC, the command itself is returned, e.g. add.
//
// Should not be called if the current command is C_RETURN.
func (p *parser) Arg1() string {
	splits := strings.Split(p.currentCommand, " ")
	switch p.CommandType() {
	case C_ARITHMETIC:
		return splits[0]
	default:
		return splits[1]
	}
}

// Returns the second argument of the current command.
// Should be called only if the current command is C_PUSH, C_POP, C_FUNCTION, C_CALL.
func (p *parser) Arg2() string {
	return strings.Split(p.currentCommand, " ")[2]
}

// tryFindNextCommand iteratively reads lines until it finds one which contains a command, or until no lines remain.
// If a command is found, it is set to be the next command (not the current), and p.hasMore is set to true.
// If a command is not found, p.hasMore is set to false.
func (p *parser) tryFindNextCommand() {
	for p.hasMore = p.scanner.Scan(); p.hasMore; p.hasMore = p.scanner.Scan() {
		stripped := removeWhitespace(stripComment(p.scanner.Text()))

		if stripped != "" {
			p.nextCommand = stripped
			return
		}
	}
}

func stripComment(s string) string {
	if idx := strings.Index(s, "//"); idx != -1 {
		return s[:idx]
	}
	return s
}

func removeWhitespace(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "\t", "")
}
