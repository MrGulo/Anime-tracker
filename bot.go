package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Константы
const ItemsPerPage = 5

// --- СТРУКТУРЫ ---
type User struct {
	ID int `json:"id"`
}
type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Data    string   `json:"data"`
	Message *Message `json:"message"`
}
type TelegramResponse struct {
	OK     bool     `json:"ok"`
	Result []Update `json:"result"`
}
type Update struct {
	UpdateId      int            `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}
type Message struct {
	ID   int    `json:"message_id"`
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
}
type Chat struct {
	ID int `json:"id"`
}

// --- БАЗОВЫЕ ФУНКЦИИ ---

// Универсальная функция для любых запросов к Telegram API
func callTelegramAPI(botToken string, method string, params url.Values) error {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/%s", botToken, method)
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := http.Get(endpoint)
	if err != nil {
		return fmt.Errorf("ошибка сети: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка API Telegram %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func getTopAnime(db *sql.DB) string {
	rows, err := db.Query("SELECT title, rating FROM anime ORDER BY rating DESC LIMIT $1", ItemsPerPage)
	if err != nil {
		log.Println("Ошибка БД в getTopAnime:", err)
		return "Ошибка при получении топа."
	}
	defer rows.Close()

	botMessage := "Топ 5-аниме:\n"
	for rows.Next() {
		var anime Anime
		err = rows.Scan(&anime.Title, &anime.Rating)
		if err != nil {
			log.Println("Ошибка чтения строки:", err)
			continue
		}
		botMessage += fmt.Sprintf("- %s (%.1f)\n", anime.Title, anime.Rating)
	}
	return botMessage
}

// --- МАРШРУТИЗАЦИЯ ---

// Обработка текстовых сообщений
func handleMessage(db *sql.DB, botToken string, message *Message) {
	chatID := strconv.Itoa(message.Chat.ID)
	text := message.Text

	var botMessage string
	var replyMarkup string

	parts := strings.SplitN(text, " ", 2)
	command := parts[0]

	switch command {
	case "/start":
		botMessage = "Привет! Я твой Аниме-помощник."
		replyMarkup = `{"inline_keyboard": [[{"text": "Узнать Топ-5", "callback_data": "btn_top"}]]}`
	case "/list":
		botMessage = getTopAnime(db)
	case "/search":
		if len(parts) == 2 {
			botMessage = "Вот то что ты искал:\n"
			searchQuery := parts[1]

			rows, err := db.Query("SELECT title, rating FROM anime WHERE title ILIKE '%' || $1 || '%' LIMIT $2 OFFSET 0", searchQuery, ItemsPerPage)
			if err != nil {
				log.Println("Ошибка БД при поиске:", err)
				return
			}
			foundCount := 0
			for rows.Next() {
				var anime Anime
				if err := rows.Scan(&anime.Title, &anime.Rating); err == nil {
					botMessage += fmt.Sprintf("- %s (%.1f)\n", anime.Title, anime.Rating)
					foundCount++
				}
			}
			rows.Close()

			if foundCount == 0 {
				botMessage = "Не получилось найти ваше аниме, возможно вы допустили опечатку."
			} else {
				replyMarkup = fmt.Sprintf(`{"inline_keyboard": [[{"text": "Вперед ▶️", "callback_data": "page_1_%s"}]]}`, searchQuery)
			}
		} else {
			botMessage = "Уточните название после команды, например: /search Наруто"
		}
	default:
		botMessage = "Я пока не понимаю эту команду."
	}

	params := url.Values{}
	params.Add("chat_id", chatID)
	params.Add("text", botMessage)
	if replyMarkup != "" {
		params.Add("reply_markup", replyMarkup)
	}

	if err := callTelegramAPI(botToken, "sendMessage", params); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// Обработка кнопок (интерактив)
func handleCallback(db *sql.DB, botToken string, callback *CallbackQuery) {
	chatID := strconv.Itoa(callback.From.ID)
	messageID := strconv.Itoa(callback.Message.ID)
	data := callback.Data

	// Гарантированно убираем прогрузку с кнопки при выходе из функции
	defer func() {
		params := url.Values{}
		params.Add("callback_query_id", callback.ID)
		_ = callTelegramAPI(botToken, "answerCallbackQuery", params)
	}()

	if data == "btn_top" {
		params := url.Values{}
		params.Add("chat_id", chatID)
		params.Add("text", getTopAnime(db))
		if err := callTelegramAPI(botToken, "sendMessage", params); err != nil {
			log.Println("Ошибка отправки Топ-5:", err)
		}
		return
	}

	if strings.HasPrefix(data, "page_") {
		partsData := strings.SplitN(data, "_", 3)
		pageNumber, err := strconv.Atoi(partsData[1])
		if err != nil {
			log.Println("Ошибка конвертации страницы:", err)
			return
		}

		searchQuery := partsData[2]
		dbOffset := pageNumber * ItemsPerPage

		rows, err := db.Query("SELECT title, rating FROM anime WHERE title ILIKE '%' || $1 || '%' LIMIT $2 OFFSET $3", searchQuery, ItemsPerPage, dbOffset)
		if err != nil {
			log.Println("Ошибка БД при пагинации:", err)
			return
		}

		botMessage := ""
		foundCount := 0
		for rows.Next() {
			var anime Anime
			if err := rows.Scan(&anime.Title, &anime.Rating); err == nil {
				botMessage += fmt.Sprintf("- %s (%.1f)\n", anime.Title, anime.Rating)
				foundCount++
			}
		}
		rows.Close()

		var replyMarkup string
		if pageNumber > 0 && foundCount == ItemsPerPage {
			replyMarkup = fmt.Sprintf(`{"inline_keyboard": [[{"text": "◀️ Назад", "callback_data": "page_%d_%s"},{"text": "Вперед ▶️", "callback_data": "page_%d_%s"}]]}`, pageNumber-1, searchQuery, pageNumber+1, searchQuery)
		} else if pageNumber > 0 && foundCount < ItemsPerPage {
			replyMarkup = fmt.Sprintf(`{"inline_keyboard": [[{"text": "◀️ Назад", "callback_data": "page_%d_%s"}]]}`, pageNumber-1, searchQuery)
		} else if pageNumber == 0 && foundCount == ItemsPerPage {
			replyMarkup = fmt.Sprintf(`{"inline_keyboard": [[{"text": "Вперед ▶️", "callback_data": "page_%d_%s"}]]}`, pageNumber+1, searchQuery)
		}

		params := url.Values{}
		params.Add("chat_id", chatID)
		params.Add("message_id", messageID)
		params.Add("text", botMessage)
		if replyMarkup != "" {
			params.Add("reply_markup", replyMarkup)
		}

		if err := callTelegramAPI(botToken, "editMessageText", params); err != nil {
			log.Println("Ошибка редактирования сообщения:", err)
		}
	} else {
		log.Println("Неизвестная кнопка:", data)
	}
}

// --- ГЛАВНЫЙ ЦИКЛ ---
func startBot(db *sql.DB) {
	// Достаем токен из системы
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("ОШИБКА: Токен не найден. Задайте переменную окружения TELEGRAM_BOT_TOKEN")
	}

	var offset = 0

	for {
		params := url.Values{}
		params.Add("offset", strconv.Itoa(offset))

		endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?%s", botToken, params.Encode())

		resp, err := http.Get(endpoint)
		if err != nil {
			log.Println("Ошибка соединения с Telegram:", err)
			time.Sleep(2 * time.Second)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Println("Ошибка чтения ответа:", err)
			continue
		}

		var telegramResponses TelegramResponse
		err = json.Unmarshal(body, &telegramResponses)
		if err != nil {
			log.Println("Ошибка парсинга JSON:", err)
			continue
		}

		// Распределяем задачи
		for _, update := range telegramResponses.Result {
			if update.Message != nil {
				handleMessage(db, botToken, update.Message)
			} else if update.CallbackQuery != nil {
				handleCallback(db, botToken, update.CallbackQuery)
			}
			offset = update.UpdateId + 1
		}
	}
}
