package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func StartBot() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	yourChatIDStr := os.Getenv("TELEGRAM_CHAT_ID")

	if botToken == "" || yourChatIDStr == "" {
		log.Fatal("Переменные TELEGRAM_BOT_TOKEN или TELEGRAM_CHAT_ID не заданы")
	}

	yourChatID, err := strconv.ParseInt(yourChatIDStr, 10, 64)
	if err != nil {
		log.Fatal("Неверный TELEGRAM_CHAT_ID:", err)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Бот успешно запущен. Работает только для chatID: %d", yourChatID)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Ошибка получения обновлений:", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Разрешён только твой Telegram ID
		if update.Message.Chat.ID != yourChatID {
			continue
		}

		switch update.Message.Text {
		case "/start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Привет! Я бот-шутник. Команды:\n"+
					"/stat - Показать количество переходов\n"+
					"/visits - Список последних переходов")
			bot.Send(msg)

		case "/stat":
			var count int
			err := DB.QueryRow("SELECT COUNT(*) FROM visits").Scan(&count)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении статистики"))
				continue
			}
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Всего переходов: %d", count)))

		case "/visits":
			rows, err := DB.Query("SELECT ip, user_agent, visited_at FROM visits ORDER BY visited_at DESC LIMIT 10")
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении данных"))
				continue
			}
			defer rows.Close()

			var msg string
			for rows.Next() {
				var ip, userAgent string
				var visitedAt time.Time

				scanErr := rows.Scan(&ip, &userAgent, &visitedAt)
				if scanErr != nil {
					continue
				}

				msg += fmt.Sprintf("🕒 %s\n🌐 %s\n📱 %s\n\n",
					visitedAt.Format("02.01.2006 15:04:05"), ip, shortenUserAgent(userAgent))
			}

			if msg == "" {
				msg = "Нет данных."
			}

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))

		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Используй /stat или /visits"))
		}
	}
}

// shortenUserAgent сокращает user-agent для удобства чтения
func shortenUserAgent(ua string) string {
	if len(ua) > 60 {
		return ua[:60] + "..."
	}
	return ua
}
