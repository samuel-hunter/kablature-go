package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func InterpretTablature(w io.Writer, r io.Reader) error {
	parser := NewParser(NewScanner(r))
	symbols := []Symbol{}

	for {
		symbol, err := parser.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		symbols = append(symbols, symbol)
	}

	return DrawTablature(w, symbols)
}

func writeToServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	file, err := os.Open("song1.tab")
	if err != nil {
		panic(err)
	}

	err = InterpretTablature(w, file)
	if err != nil {
		panic(err)
	}
}

func main() {
	http.Handle("/", http.HandlerFunc(writeToServer))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
