package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"password-manager/db"
	"password-manager/handlers"
)

func main() {
	// Инициализация БД
	database, err := db.InitDB("./passwords.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	e := echo.New()

	// Middleware логирования
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Logger().Infof("REQUEST: %s %s", c.Request().Method, c.Request().URL)
			return next(c)
		}
	})

	// Роуты
	e.GET("/passwords", handlers.GetPasswords)
	e.GET("/passwords/:id", handlers.GetPassword)
	e.POST("/passwords", handlers.CreatePassword)
	e.PUT("/passwords/:id", handlers.UpdatePassword)
	e.DELETE("/passwords/:id", handlers.DeletePassword)
	e.POST("/passwords/:id/copy", handlers.CopyPassword)

	log.Println("Сервер запущен на :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
