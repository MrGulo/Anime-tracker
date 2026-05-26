package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func GetAnimeHandler(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")

	limitInt, err := strconv.Atoi(limitStr)
	if err != nil || limitInt <= 0 {
		limitInt = 10
	}

	pageInt, err := strconv.Atoi(pageStr)
	if err != nil || pageInt <= 0 {
		pageInt = 1
	}
	fmt.Printf("Пользователь запросил лимит: %d, страница: %d\n", limitInt, pageInt)

	offset := (pageInt - 1) * limitInt

	rows, err := db.Query("SELECT * FROM anime ORDER BY id ASC LIMIT $1 OFFSET $2", limitInt, offset)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var animes []Anime

	for rows.Next() {
		var a Anime
		err := rows.Scan(&a.ID, &a.Title, &a.Status, &a.Rating)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		animes = append(animes, a)
	}
	json.NewEncoder(w).Encode(animes)
}

func addAnimeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	var newAnime Anime
	json.NewDecoder(r.Body).Decode(&newAnime)

	query := "INSERT INTO anime (title, status, rating) VALUES ($1, $2, $3) RETURNING id"
	err := db.QueryRow(query, newAnime.Title, newAnime.Status, newAnime.Rating).Scan(&newAnime.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(newAnime)
}

func searchAnimeHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	query := "SELECT id,title,status,rating FROM anime WHERE id = $1"
	var a Anime

	err = db.QueryRow(query, id).Scan(&a.ID, &a.Title, &a.Status, &a.Rating)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(a)
}

func deleteAnimeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Wrong ID", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM anime WHERE id = $1"

	_, err = db.Exec(query, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updateAnimeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		return
	}
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "Wrong ID", http.StatusBadRequest)
		return
	}

	var updateAnime Anime
	json.NewDecoder(r.Body).Decode(&updateAnime)

	query := "UPDATE anime SET title = $1, status = $2, rating = $3 WHERE id = $4"
	_, err = db.Exec(query, updateAnime.Title, updateAnime.Status, updateAnime.Rating, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
