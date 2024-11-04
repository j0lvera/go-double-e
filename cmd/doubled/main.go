package main

import (
	"log"
	"net/http"
)

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Hello, World!"))
		if err != nil {
			return
		}
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Println("Server is running on port 8080")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
