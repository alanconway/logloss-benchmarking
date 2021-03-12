package main

import "testing"

type nullWriter struct{}

func (nullWriter) WriteByte(byte) error { return nil }

func BenchmarkWriteString(b *testing.B) {
	r := NewRand()
	w := nullWriter{}
	r.WriteString(w, 1024)
}
