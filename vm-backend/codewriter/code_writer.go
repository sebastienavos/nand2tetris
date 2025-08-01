package codewriter

import "io"

type CodeWriter struct {
	w io.WriteCloser
}

func NewCodeWriter(w io.WriteCloser) CodeWriter {
	return CodeWriter{w}
}

func (c CodeWriter) SetFilename(filename string)                 {}
func (c CodeWriter) WriteArithmetic(cmd string)                  {}
func (c CodeWriter) WritePushPop(cmd, segment string, index int) {}

func (c CodeWriter) Close() { c.w.Close() }
