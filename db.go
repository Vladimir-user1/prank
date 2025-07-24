package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	var err error
	dbURL := os.Getenv("DATABASE_URL")

	DB, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("БД не отвечает:", err)
	}

	log.Println("Успешное подключение к БД")
}
