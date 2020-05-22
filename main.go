package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	fmt.Println("run")

	run()
}

func run() {
	r := mux.NewRouter()

	// html
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/html")
		w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.ServeFile(w, r, "./www/index.html")
	})
	r.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/html")
		w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.ServeFile(w, r, "./www/success.html")
	})

	// rest
	r.HandleFunc("/webhook", webhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(":"+port, r))
}

func webhook(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, err)
	}

	result := struct {
		Body string
	}{
		Body: string(body),
	}

	data, err := json.Marshal(&result)
	if err != nil {
		writeError(w, err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)

	w.Write([]byte(err.Error()))
}
