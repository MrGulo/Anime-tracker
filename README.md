# Anime Tracker (Bot & API)

A robust, highly concurrent Go application for tracking and managing anime lists. This project elegantly combines a Telegram Bot interface, a RESTful API, and an autonomous background worker within a single application, demonstrating advanced knowledge of Go's concurrency model (Goroutines).

---

## Tech Stack & Architecture

* **Language:** Go (Golang)
* **Database:** PostgreSQL (utilizing `ON CONFLICT DO NOTHING` for data consistency)
* **External APIs:** * Telegram Bot API (for user interaction)
  * Shikimori Public API (for automatic data fetching)
* **Concurrency:** Native Goroutines and Channels orchestrating three independent subsystems simultaneously.

---

## Core Subsystems

This application runs three independent components concurrently:

### 1. 🤖 Telegram Bot Interface (`bot.go`)
Provides a user-friendly way to interact with the database directly from the messenger.
* **Smart Pagination:** Implements dynamic pagination for search results using `editMessageText` and `InlineKeyboardMarkup`, preventing chat history clutter.
* **Commands:** Supports seamless routing for `/start`, `/list`, and `/search`.

### 2. 🌐 RESTful API (`handles.go`)
A full-fledged CRUD interface built strictly with the standard `net/http` package.
* **External Management:** Allows external clients to Create, Read, Update, and Delete anime entries in the database.
* **Pagination Support:** Handles large datasets efficiently via query parameters (e.g., `?limit=10&page=2`).

### 3. ⚙️ Autonomous Background Worker (`work.go`)
A background Goroutine dedicated to keeping the local database synchronized.
* **Daily Sync:** Automatically sends scheduled requests to the public Shikimori API.
* **JSON Parsing:** Processes incoming data and seamlessly populates the local database with new titles, entirely isolated from user-facing operations.

---

## Getting Started (Local Development)

### Prerequisites
* Go 1.20+ installed.
* A PostgreSQL instance running.
* A Telegram Bot Token (from [@BotFather](https://t.me/BotFather)).

### Installation

1. **Clone the repository:**
   ```bash
   git clone [https://github.com/MrGulo/anime-tracker.git](https://github.com/MrGulo/anime-tracker.git)
   cd anime-tracker
   ```

2. **Environment Setup:**
   Rename the provided `.env.example` file to `.env` and fill in your actual database credentials and Telegram token:
   ```bash
   cp .env.example .env
   ```

3. **Install Dependencies:**
   ```bash
   go mod download
   ```

4. **Run the Application:**
   ```bash
   go run main.go
   ```
   *The bot will come online, the REST API will start listening for requests, and the background worker will begin its initial sync.*

---
*Developed to master API integrations, concurrent architecture, and seamless bot interactions in pure Go.*
