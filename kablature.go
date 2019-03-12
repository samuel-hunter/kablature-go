package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var inputFile string
var outputFile string

func InterpretFile(w io.Writer) error {
	file, err := os.Open(inputFile)
	if err != nil {
		return err
	}

	parser := NewParser(file)
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
	flag.StringVar(&inputFile, "i", "", "Input file")
	flag.StringVar(&outputFile, "o", "out.svg", "Output file")
}

func main() {
	flag.Parse()

	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Input file is not specified.")
		os.Exit(1)
	}

	out, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}

	InterpretFile(out)
}
