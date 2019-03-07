package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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

	fmt.Println(symbols)

	return DrawTablature(w, symbols)
}

func writeToServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")

	err := InterpretTablature(w, strings.NewReader("4 e 1 c 2 (c e g)"))
	// err := DrawTablature(w, []Symbol{
	// 	Note{length: 4, dotted: false, pitch: 2},
	// 	Note{length: 1, dotted: false, pitch: 0},
	// 	Chord{length: 2, dotted: false, pitches: []byte{0, 2, 4}},
	// })

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
