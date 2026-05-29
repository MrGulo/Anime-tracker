package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

type Shikimori struct {
	ID       int    `json:"id"`
	TitleEng string `json:"name"`
	TitleRus string `json:"russian"`
	Rating   string `json:"score"`
	Status   string `json:"status"`
}

func fetchAndSaveAnime() {
	page := 1
	getURL := fmt.Sprintf("https://shikimori.one/api/animes?limit=50&page=%d", page)

	resp, err := http.Get(getURL)
	if err != nil {
		fmt.Println("Ошибка подключения к shikimori", err)
		return
	}

	var animes []Shikimori
	err = json.NewDecoder(resp.Body).Decode(&animes)
	resp.Body.Close()
	if err != nil {
		fmt.Println("Ошибка чтения JSON в воркере", err)
		return
	}

	var addedTitles []string

	for _, anime := range animes {
		ratingFloat, _ := strconv.ParseFloat(anime.Rating, 64)

		res, err := db.Exec(`INSERT INTO anime (title, status, rating) VALUES ($1, $2, $3)
			 ON CONFLICT (title) DO NOTHING`, anime.TitleRus, "В планах", ratingFloat)
		if err != nil {
			fmt.Println(err)
		}

		rows, _ := res.RowsAffected()

		if rows == 1 {
			addedTitles = append(addedTitles, anime.TitleRus)
		}
	}

	if len(addedTitles) > 0 {
		fmt.Printf("Лог воркера [%s]: Синхронизация успешна!\n", time.Now().Format("15:04:02"))
		fmt.Printf("Добавлено новых тайтлов: %d\n", len(addedTitles))
		fmt.Println("Номер страницы:", page)
		fmt.Println("Список новинок:")

		for i, title := range addedTitles {
			fmt.Printf("%d. %s\n", i+1, title)
		}

		fmt.Println("--------------------------------------------------")
	} else {
		fmt.Printf("Лог воркера [%s]: Новых аниме на Шикимори не появилось.\n", time.Now().Format("15:04:02"))
	}
	page = page + 1
	time.Sleep(1 * time.Second)
	//if len(animes) == 0 {
	//	break
	//}

}

func startWorker() {
	ticker := time.NewTicker(24 * time.Hour)

	fetchAndSaveAnime()

	for range ticker.C {
		fetchAndSaveAnime()
	}
}
