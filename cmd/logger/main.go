package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/pmoogi-redhat/logloss-benchmarking/pkg/generator"
)

var (
	length         = flag.Int("length", 128, "Mean length of lines in bytes")
	stddev         = flag.Float64("stddev", 32, "Standard deviation for random line lengths")
	distribution   = flag.String("distribution", "normal", "Distribution for random line lengths: 'fixed', 'normal'")
	rate           = flag.Float64("rate", 0, "Lines per second, 0 means unlimited")
	id             = flag.String("id", "", "Identifying string to include in lines")
	reportInterval = flag.Duration("report-time", 0, "Time interval between reports, 0 means no report till exit")
	output         = flag.String("output", "stdout", "Log destination: 'stdout', 'stderr' or file name")
	report         = flag.String("report", "stderr", "Report destination: 'stdout', 'stderr' or file name")
	duration       = flag.Duration("time", 0, "Exit after this duration has elapsed, 0 unlimited")
	total          = flag.Int64("total", 0, "Exit after this many MB of data are logged, 0 unlimited")
	seed           = flag.Int64("seed", 0, "Random seed for repeatable runs, 0 means randomize the run")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage %v [FLAGS]
Print lines to simulate logging, print a statistical summary at intervals and on completion.
Flags:
`, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(*seed))
	gen := generator.Generator{
		Rate:           *rate,
		Rand:           r,
		LogWriter:      outputWriter(*output),
		ReportWriter:   outputWriter(*report),
		ReportInterval: *reportInterval,
		Id:             *id,
	}
	switch *distribution {
	case "fixed":
		gen.Length = func() int { return *length }
	case "normal":
		gen.Length = func() int { return int(r.NormFloat64()*(*stddev) + float64(*length)) }
	}
	exitIf(gen.Run(*duration, *total*1024*1024))
	exitIf(gen.Report()) // Always generate a final report.
}

func exitIf(err error) {
	if err != nil {
		log.Fatalf("fatal: %s", err)
	}
}

func outputWriter(name string) io.Writer {
	switch name {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	default:
		w, err := os.Create(name)
		exitIf(err)
		return w
	}
}
