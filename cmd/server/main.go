package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "strings"

    "password-manager/internal/app"
    "password-manager/internal/app/endpoint"

    "github.com/joho/godotenv"
    "github.com/labstack/echo/v4"
)

func main() {
    // Загружаем переменные окружения из .env
    if err := godotenv.Load(); err != nil {
        log.Fatal("Ошибка загрузки .env файла")
    }

    // Проверяем наличие ключа шифрования
    if os.Getenv("ENCRYPTION_KEY") == "" {
        log.Fatal("ENCRYPTION_KEY не задан! Добавьте его в .env")
    }

    // Проверяем наличие мастер-пароля
    if os.Getenv("MASTER_PASSWORD") == "" {
        log.Fatal("MASTER_PASSWORD не задан! Добавьте его в .env")
    }

    // Запрос мастер-пароля
    fmt.Print("Введите мастер-пароль: ")
    input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
    input = strings.TrimSpace(input)

    if input != os.Getenv("MASTER_PASSWORD") {
        log.Fatal("Неверный мастер-пароль")
    }

    e := echo.New()

    // Middleware логирования
    e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            c.Logger().Infof("REQUEST: %s %s", c.Request().Method, c.Request().URL)
            return next(c)
        }
    })

    // Инициализация приложения
    appInstance := app.InitApp(e)
    defer appInstance.DB.Close()

    // Регистрируем маршруты
    endpoint.RegisterRoutes(e, appInstance)

    log.Println("Сервер запущен на :8080")
    e.Logger.Fatal(e.Start(":8080"))
}
