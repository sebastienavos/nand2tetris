package main

import (
	"log"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <input.vm | input_directory>", os.Args[0])
	}

	vmFiles := statFiles(os.Args[1])

	out, err := os.Create("Prog.asm")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range vmFiles {
		translate(f, out)
	}
}

func translate(f string, out *os.File) {
	panic("unimplemented")
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
