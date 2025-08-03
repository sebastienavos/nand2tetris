package parser

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

type CommandType = string

const (
	C_ARITHMETIC CommandType = "C_ARITHMETIC"
	C_PUSH       CommandType = "C_PUSH"
	C_POP        CommandType = "C_POP"
	C_LABEL      CommandType = "C_LABEL"
	C_GOTO       CommandType = "C_GOTO"
	C_IF_GOTO    CommandType = "C_IF"
	C_FUNCTION   CommandType = "C_FUNCTION"
	C_RETURN     CommandType = "C_RETURN"
	C_CALL       CommandType = "C_CALL"
)

var cmdMapping = map[string]CommandType{
	"add":      C_ARITHMETIC,
	"sub":      C_ARITHMETIC,
	"neg":      C_ARITHMETIC,
	"eq":       C_ARITHMETIC,
	"gt":       C_ARITHMETIC,
	"lt":       C_ARITHMETIC,
	"and":      C_ARITHMETIC,
	"or":       C_ARITHMETIC,
	"not":      C_ARITHMETIC,
	"push":     C_PUSH,
	"pop":      C_POP,
	"label":    C_LABEL,
	"goto":     C_GOTO,
	"if-goto":  C_IF_GOTO,
	"function": C_FUNCTION,
	"return":   C_RETURN,
	"call":     C_CALL,
}

type parser struct {
	scanner *bufio.Scanner
	// hasMore means next is valid
	hasMore        bool
	currentCommand []string
	nextCommand    []string
}

func NewParserFromFilename(fileName string) *parser {
	file, err := os.Open(fileName)
	if err != nil {
		panic("could not open file" + fileName)
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
	cmd, ok := cmdMapping[p.currentCommand[0]]
	if !ok {
		panic("command not recognised: " + p.currentCommand[0])
	}
	return cmd
}

// Returns the first argument of the current command.
// In the case of C_ARITHMETIC, the command itself is returned, e.g. add.
//
// Should not be called if the current command is C_RETURN.
func (p *parser) Arg1() string {
	switch p.CommandType() {
	case C_ARITHMETIC:
		return p.currentCommand[0]
	default:
		return p.currentCommand[1]
	}
}

// Returns the second argument of the current command.
// Should be called only if the current command is C_PUSH, C_POP, C_FUNCTION, C_CALL.
func (p *parser) Arg2() int {
	i, err := strconv.Atoi(p.currentCommand[2])
	if err != nil {
		panic("oh bugger")
	}
	return i
}

// tryFindNextCommand iteratively reads lines until it finds one which contains a command, or until no lines remain.
// If a command is found, it is set to be the next command (not the current), and p.hasMore is set to true.
// If a command is not found, p.hasMore is set to false.
func (p *parser) tryFindNextCommand() {
	for p.hasMore = p.scanner.Scan(); p.hasMore; p.hasMore = p.scanner.Scan() {
		fields := strings.Fields(stripComment(p.scanner.Text()))
		if len(fields) > 0 {
			p.nextCommand = fields
			return
		}
	}
}

func stripComment(s string) string {
	noComment, _, _ := strings.Cut(s, "//")
	return noComment
}
