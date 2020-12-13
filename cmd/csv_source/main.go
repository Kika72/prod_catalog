package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func main() {
	m := http.NewServeMux()
	m.HandleFunc("/csv", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("idx")
		if len(key) == 0 {
			log.Println("file number is missing")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("file number is missing"))
			return
		}

		idx, err := strconv.Atoi(key)
		if err != nil {
			log.Println("bad file number")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad file number"))
			return
		}
		idx++

		w.Header().Set("Content-Type", "plain/text")
		w.WriteHeader(http.StatusOK)
		for i := 1; i <= 10; i++ {
			w.Write([]byte(fmt.Sprintf("name %d;%d\n",
				i*10*idx,
				idx*i,
			),
			))
		}
	})

	log.Println("Starting...")
	if err := http.ListenAndServe("0.0.0.0:3000", m); err != nil {
		log.Fatal(err)
	}
}
