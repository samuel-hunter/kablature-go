package main

import (
	"log"
	"net/http"
)

func hello(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")

	DrawTablature(w, []Symbol{
		Note{length: 1, dotted: false, pitch: 0},
		Note{length: 1, dotted: false, pitch: 1},
		Note{length: 2, dotted: false, pitch: 2},
		Note{length: 4, dotted: false, pitch: 3},
		Note{length: 8, dotted: false, pitch: 4},
	})
}

func main() {
	http.Handle("/", http.HandlerFunc(hello))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
