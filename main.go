package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {

	// Загрузка .env только локально
	if _, ok := os.LookupEnv("RENDER"); !ok {
		err := godotenv.Load()
		if err != nil {
			log.Println(" .env файл не загружен (локальный запуск)")
		} else {
			log.Println(" .env файл загружен")
		}
	}

	InitDB()

	go startBot()

	http.HandleFunc("/", prankHandler)
	http.HandleFunc("/count", countHandler)

	port := os.Getenv("PORT") // For Render
	if port == "" {
		port = "10000"
	}
	fmt.Println("Сервер запущен на порту", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func startBot() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")

	if botToken == "" || chatIDStr == "" {
		log.Println("TELEGRAM_BOT_TOKEN или TELEGRAM_CHAT_ID не указаны")
		return
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Println("Неверный TELEGRAM_CHAT_ID:", err)
		return
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Println("Ошибка запуска бота:", err)
		return
	}

	log.Println("Telegram-бот запущен")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Println("Ошибка получения обновлений:", err)
		return
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Chat.ID != chatID {
			continue // Игнорировать других пользователей
		}

		if update.Message.Text == "/count" {
			var count int
			err := DB.QueryRow("SELECT COUNT(*) FROM visits").Scan(&count)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Ошибка БД"))
				continue
			}

			msg := fmt.Sprintf("👀 Всего переходов: %d", count)
			bot.Send(tgbotapi.NewMessage(chatID, msg))
		}
	}
}

func prankHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	// Cut port from IP
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	_, err := DB.Exec(`insert into visits (ip, user_agent, visited_at) values ($1, $2, $3)`,
		ip, userAgent, time.Now())
	if err != nil {
		log.Println("Ошибка записи в БД:", err)
	}

	fmt.Fprintf(w, `
<!DOCTYPE html>
		<html lang="ru">
		<head>
			<meta charset="UTF-8">
			<title>Сюрприз</title>
			<style>
				body { font-family: sans-serif; text-align: center; margin-top: 20%%; background-color: #fafafa; }
				h1 { color: #e91e63; }
			</style>
		</head>
		<body>
			<h1>Ты думал будешь смотреть на 🍆 в OnlyFans!? Поздравляю с 1 апреля 2026, тебя наебали </h1>
		</body>
		</html>
	`)
}

func countHandler(w http.ResponseWriter, r *http.Request) {
	var count int
	err := DB.QueryRow("Select count(*) from visits").Scan(&count)
	if err != nil {
		log.Println("Ошибка подсчета:", err)
		w.WriteHeader(500)
		return
	}

	fmt.Fprintf(w, "Всего переходов: %d\n", count)
}
