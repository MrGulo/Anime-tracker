package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	initDB()
	createTable()

	go startWorker()
	go startBot(db)

	http.HandleFunc("/anime", GetAnimeHandler)
	http.HandleFunc("/add", addAnimeHandler)
	http.HandleFunc("/search", searchAnimeHandler)
	http.HandleFunc("/delete", deleteAnimeHandler)
	http.HandleFunc("/update", updateAnimeHandler)

	fmt.Println("Listening on port 8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
