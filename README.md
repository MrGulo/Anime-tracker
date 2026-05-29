# 🍙 Anime Tracker Microservice

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-4169E1?logo=postgresql)](https://www.postgresql.org/)
[![Telegram API](https://img.shields.io/badge/Telegram_Bot-API-2CA5E0?logo=telegram)](https://core.telegram.org/bots)

A comprehensive microservice for tracking and managing anime lists. This project demonstrates skills in building a reliable backend architecture, utilizing concurrency (Goroutines), and integrating with third-party APIs.

## 🚀 Features

The project consists of three independent components running concurrently within a single application:

1. **Telegram Bot Interface (`bot.go`)**
   - User-friendly database search directly via the messenger.
   - Dynamic pagination for search results without cluttering the chat history (utilizing `editMessageText` and `InlineKeyboardMarkup`).
   - Command routing (`/start`, `/list`, `/search`).

2. **RESTful API (`handles.go`)**
   - Full-fledged CRUD interface using the standard `net/http` package for external database management (POST, GET, PUT, DELETE).
   - Pagination support via query parameters (`?limit=` and `?page=`).

3. **Background Worker (`work.go`)**
   - A background process (Goroutine) that automatically synchronizes the database.
   - Sends daily requests to the public **Shikimori** API, parses JSON responses, and seamlessly populates the local database with new titles (preventing duplicates using `ON CONFLICT DO NOTHING`).

## 🛠 Tech Stack

* **Language:** Go (Golang)
* **Database:** PostgreSQL
* **DB Driver:** `github.com/lib/pq`
* **Architecture:** Monolithic application with concurrent sub-services, Stateless Bot logic.

## ⚙️ Quick Start

### 1. Clone the repository
```bash
git clone https://github.com/MrGulo/Anime-tracker.git
cd Anime-tracker
