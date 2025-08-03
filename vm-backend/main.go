package main

import (
	"log"
	"os"
	"path/filepath"
	"vm/codewriter"
	"vm/parser"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: go run main.go <input.vm | input_directory> <outfile>")
	}

	vmFiles := statFiles(os.Args[1])

	out, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	cw := codewriter.NewCodeWriter(out)
	defer cw.Close()

	for _, f := range vmFiles {
		cw.SetFilename(f)
		p := parser.NewParserFromFilename(f)

		for p.HasMoreCommands() {
			p.Advance()
			switch p.CommandType() {
			case parser.C_ARITHMETIC:
				cw.WriteArithmetic(p.Arg1())
			case parser.C_POP:
				cw.WritePushPop("pop", p.Arg1(), p.Arg2())
			case parser.C_PUSH:
				cw.WritePushPop("push", p.Arg1(), p.Arg2())
			case parser.C_LABEL:
				cw.WriteLabel(p.Arg1())
			case parser.C_GOTO:
				cw.WriteGoto(p.Arg1())
			case parser.C_IF:
				cw.WriteIfGoto(p.Arg1())
			default:
				panic("crikey")
			}
		}
	}
}

func statFiles(in string) []string {
	info, err := os.Stat(in)
	if err != nil {
		log.Fatal(err)
	}

	if info.IsDir() {
		vmFiles := []string{}

		entries, err := os.ReadDir(in)
		if err != nil {
			log.Fatal(err)
		}
		for _, e := range entries {
			if e.IsDir() {
				log.Printf("skipping subdirectory %v", e.Name())
				continue
			}

			if filepath.Ext(e.Name()) != ".vm" {
				log.Printf("skipping file with wrong extension: %v", e.Name())
			}

			vmFiles = append(vmFiles, e.Name())
		}
		return vmFiles
	}

	if filepath.Ext(in) != ".vm" {
		log.Fatalf("Expected a .vm file, got '%s'", in)
	}
	return []string{info.Name()}
}
