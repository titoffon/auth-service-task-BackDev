package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/titoffon/auth-service-task-BackDev/internal/db"
	"github.com/titoffon/auth-service-task-BackDev/internal/handlers"
)

func main() {
	
	ctx := context.Background()

	err := godotenv.Load()
  	if err != nil {
    	log.Printf("Ошибка при загрузке файла .env: %v", err)
        os.Exit(1)

	}

    // Соединение с базой данных
	pool, err := db.InitDB(ctx)
    if err != nil {
        log.Printf("Не удалось подключиться к базе данных: %v", err)
        os.Exit(1)
	}
	defer pool.Pool.Close()
	log.Println("Соединение с базой данных установлено")

    // Инициализация роутера
    r := gin.Default()

	// Создание экземпляра AuthHandler
	authHandlers := handlers.NewAuthHandler(pool)

    r.POST("/auth/generate-tokens", authHandlers.GenerateTokens)
    r.POST("/auth/refresh-tokens", authHandlers.RefreshTokens)

    r.Run(os.Getenv("PORT"))
}
