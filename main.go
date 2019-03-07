package main

import (
	"log"
	"net/http"
)

func hello(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")

	err := DrawTablature(w, []Symbol{
		Note{length: 4, dotted: false, pitch: 3},
		Note{length: 1, dotted: false, pitch: 1},
		Chord{length: 2, dotted: true, pitches: []byte{
			0, 2, 4,
		}},
	})
	if err != nil {
		panic(err)
	}
}

func main() {
	http.Handle("/", http.HandlerFunc(hello))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
