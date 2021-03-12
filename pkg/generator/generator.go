package generator

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"time"
)

type Generator struct {
	LogWriter      io.Writer
	ReportWriter   io.Writer
	Length         func() int
	Id             string
	ReportInterval time.Duration
	Rate           float64
	Rand           *rand.Rand

	start        time.Time
	bytesWritten int64
	buffer       bytes.Buffer
	seq          int64
}

func (g *Generator) Run(duration time.Duration, total int64) error {
	if g.Rand == nil {
		g.Rand = rand.New(rand.NewSource(0))
	}
	interval := time.Duration(0)
	if g.Rate > 0 {
		interval = time.Duration(float64(time.Second) / g.Rate)
	}
	g.start = time.Now()
	deadline := g.start.Add(duration)
	nextReport := time.Now().Add(g.ReportInterval)
	for {
		nextWrite := time.Now().Add(interval)
		g.writeLine()
		now := time.Now()
		if (duration > 0 && now.After(deadline)) ||
			(total > 0 && g.bytesWritten > total) {
			return nil
		}
		if g.ReportInterval != 0 && now.After(nextReport) {
			g.Report()
			nextReport = time.Now().Add(g.ReportInterval)
		}
		time.Sleep(time.Until(nextWrite))
	}
}

func (g *Generator) Report() error {
	elapsed := time.Since(g.start).Seconds()
	fmt.Fprintf(g.ReportWriter, "Report %v: sequence: %v lines/sec: %v bytes written: %v bytes/sec: %v\n",
		time.Now().Format("[15:04:05.99]"), g.seq, float64(g.seq)/elapsed, g.bytesWritten, float64(g.bytesWritten)/elapsed)
	return nil
}

func (g *Generator) writeLine() {
	// IMPORTANT: generate each line in the same buffer to avoid allocations
	g.buffer.Reset()
	fmt.Fprintf(&g.buffer, "%s[%010d] ", g.Id, g.seq)
	len := g.Length()
	for len > 0 {
		v := g.Rand.Uint32()
		for i := 0; i < 4 && len > 0; i++ {
			g.buffer.WriteByte(chars256[v&0xff])
			v = v >> 8
			len--
		}
	}
	g.buffer.WriteByte('\n')
	g.LogWriter.Write(g.buffer.Bytes())
	g.bytesWritten += int64(g.buffer.Len())
	g.seq++
}

// Use base64 standard encoding chars for random strings,
const (
	base64chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	chars256    = base64chars + base64chars + base64chars + base64chars
)
