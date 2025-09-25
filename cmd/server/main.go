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
    // Load environment variables from .env
    if err := godotenv.Load(); err != nil {
        log.Fatal("Failed to load .env file")
    }

    // Check if encryption key is set
    if os.Getenv("ENCRYPTION_KEY") == "" {
        log.Fatal("ENCRYPTION_KEY is not set! Please add it to your .env file")
    }

    // Check if master password is set
    if os.Getenv("MASTER_PASSWORD") == "" {
        log.Fatal("MASTER_PASSWORD is not set! Please add it to your .env file")
    }

    // Prompt for master password
    fmt.Print("Enter master password: ")
    input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
    input = strings.TrimSpace(input)

    if input != os.Getenv("MASTER_PASSWORD") {
        log.Fatal("Incorrect master password")
    }

    e := echo.New()

    // Logging middleware
    e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            c.Logger().Infof("REQUEST: %s %s", c.Request().Method, c.Request().URL)
            return next(c)
        }
    })

    // Initialize application
    appInstance := app.InitApp(e)
    defer appInstance.DB.Close()

    // Register routes
    endpoint.RegisterRoutes(e, appInstance)

    log.Println("Server started on :8080")
    e.Logger.Fatal(e.Start(":8080"))
}
