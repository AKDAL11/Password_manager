package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var EncryptionKey []byte

func InitKey() {
	// Загружаем .env
	_ = godotenv.Load() // если файла нет, просто игнорируем

	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		log.Fatal("ENCRYPTION_KEY не задан! Добавьте его в .env")
	}

	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		log.Fatal("ENCRYPTION_KEY должен быть длиной 16, 24 или 32 байта")
	}

	EncryptionKey = []byte(key)
}
