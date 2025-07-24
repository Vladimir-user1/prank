package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Не удалось загрузить .env файл")
	}

	InitDB()

	http.HandleFunc("/", prankHandler)
	http.HandleFunc("/count", countHandler)

	port := "10000" // For Render
	fmt.Println("Сервер запущен на порту", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
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
			<h1>Ты думал будешь смотреть на 🍆 в OnlyFans!? Поздравляю с 1 апреля 2026, я тебя наебал </h1>
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
