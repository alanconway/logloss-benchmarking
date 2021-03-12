package generator

import (
	"testing"
)

type nullWriter struct{}

func (nullWriter) Write([]byte) (int, error) { return 0, nil }

func BenchmarkRun(b *testing.B) {
	gen := Generator{
		LogWriter:    nullWriter{},
		ReportWriter: nullWriter{},
		Length:       func() int { return 128 },
		Id:           "test",
	}
	gen.Run(0, 1024)
}
