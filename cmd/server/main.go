package main

import (
	"log"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok")) //cnvert string to byte then write response
}

func main() {
	http.HandleFunc("/health", healthHandler) //connects /health to healthHandler function

	log.Println("Server listening on port 8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server failed to start", err) //fatal gives timestmp error code and msg and terminates in case of error(for starter failure)
		//panic gives a long trace of error message making it easier to debug code. for unexpected states that should never happen
	}
}
