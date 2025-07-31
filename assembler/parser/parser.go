package parser

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type CommandType = string

const (
	A_COMMAND CommandType = "A_COMMAND"
	L_COMMAND CommandType = "L_COMMAND"
	C_COMMAND CommandType = "C_COMMAND"
)

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

// Returns the type of the current command:
//
//   - A_COMMAND for @Xxx where Xxx is either a symbol or a decimal number.
//   - C_COMMAND for dest=comp;jump
//   - L_COMMAND (actually, pseudo-command) for (Xxx) where Xxx is a symbol.
func (p *parser) CommandType() CommandType {
	x := p.currentCommand[0]
	switch x {
	case '@':
		return A_COMMAND
	case '(':
		return L_COMMAND
	default:
		return C_COMMAND
	}
}

// Returns the symbol or decimal Xxx of the current command @Xxx or (Xxx).
// Should be called only when commandType() is A_COMMAND or L_COMMAND.
func (p *parser) Symbol() string {
	if p.CommandType() == A_COMMAND {
		return p.currentCommand[1:]
	}
	return p.currentCommand[1 : len(p.currentCommand)-1]
}

// Returns the dest mnemonic in the current C-command (8 possibilities).
// Should be called only when commandType() is C_COMMAND.
func (p *parser) Dest() string {
	dest, _, found := strings.Cut(p.currentCommand, "=")

	if !found {
		return "null"
	}
	return dest
}

// Returns the jump mnemonic in the current C-command (8 possibilities).
// Should be called only when commandType() is C_COMMAND.
func (p *parser) Jump() string {
	_, jump, found := strings.Cut(p.currentCommand, ";")

	if !found {
		return "null"
	}
	return jump
}

// Returns the comp mnemonic in the current C-command (28 possibilities).
// Should be called only when commandType() is C_COMMAND.
func (p *parser) Comp() string {
	noJump, _, _ := strings.Cut(p.currentCommand, ";")

	dest, comp, found := strings.Cut(noJump, "=")
	if !found {
		comp = dest
	}

	return comp
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
