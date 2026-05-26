package main

import (
	_ "github.com/lib/pq"
)

type Anime struct {
	ID     int     `json:"id"`
	Title  string  `json:"title"`
	Status string  `json:"status"`
	Rating float64 `json:"rating"`
}
