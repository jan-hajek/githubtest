package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
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


	// rest
	r.HandleFunc("/test", serve)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(":"+port, r))
}

func serve(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")


	result := struct {
		Kolo string
	} {
		Kolo: "asd",
	}

	data, err := json.Marshal(&result)
	if err != nil {
		panic(err)
	}

	w.Write(data)
}