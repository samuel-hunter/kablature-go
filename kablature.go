package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

type Config struct {
	inputFile  string
	outputFile string

	BeatsPerMeasure int
	MeasuresPerTab  int
}

var GlobalConfig Config

func InterpretFile(w io.Writer, r io.Reader) error {
	parser := NewParser(r)
	var symbols []Symbol

	for {
		symbol, err := parser.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		symbols = append(symbols, symbol)
	}

	return DrawScore(w, symbols)
}

func init() {
	flag.StringVar(&GlobalConfig.inputFile, "i", "", "Input file")
	flag.StringVar(&GlobalConfig.outputFile, "o", "out.svg", "Output file")
	flag.IntVar(&GlobalConfig.BeatsPerMeasure, "b", 8, "Beats per measure")
	flag.IntVar(&GlobalConfig.MeasuresPerTab, "m", 7, "Measures per tab")
}

func main() {
	flag.Parse()

	if GlobalConfig.inputFile == "" {
		fmt.Fprintln(os.Stderr, "Input file is not specified.")
		os.Exit(1)
	}

	in, err := os.Open(GlobalConfig.inputFile)
	if err != nil {
		panic(err)
	}

	out, err := os.Create(GlobalConfig.outputFile)
	if err != nil {
		panic(err)
	}

	InterpretFile(out, in)
}
