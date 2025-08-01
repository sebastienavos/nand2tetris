package codewriter

import "io"

type codeWriter struct {
	w io.WriteCloser
}

func NewCodeWriter(w io.WriteCloser) codeWriter {
	return codeWriter{w}
}

func (c codeWriter) SetFilename(filename string)                 {}
func (c codeWriter) WriteArithmetic(cmd string)                  {}
func (c codeWriter) WritePushPop(cmd, segment string, index int) {}

func (c codeWriter) Close() { c.w.Close() }
