package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"runtime"

	"fyne.io/fyne/v2"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var encryptionKey []byte

func LoadEncryptionKey(app fyne.App) ([]byte, error) {
	if encryptionKey != nil {
		return encryptionKey, nil
	}

	// --- Android ---
	if runtime.GOOS == "android" && app != nil {
		uri := app.Storage().RootURI()
		path := uri.Path() + "/.key"

		if data, err := os.ReadFile(path); err == nil {
			key, err := base64.StdEncoding.DecodeString(string(data))
			if err != nil {
				return nil, err
			}
			encryptionKey = key
			return encryptionKey, nil
		}

		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(base64.StdEncoding.EncodeToString(key)), 0600); err != nil {
			return nil, err
		}
		encryptionKey = key
		return encryptionKey, nil
	}

	// --- Desktop ---
	_ = godotenv.Load() // пробуем загрузить, но не падаем
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if keyStr == "" {
		// генерим новый ключ
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		encoded := base64.StdEncoding.EncodeToString(key)

		// создаём .env с этим ключом
		f, err := os.Create(".env")
		if err != nil {
			return nil, err
		}
		defer f.Close()
		fmt.Fprintf(f, "ENCRYPTION_KEY=%s\n", encoded)

		encryptionKey = key
		return encryptionKey, nil
	}

	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, err
	}
	encryptionKey = key
	return encryptionKey, nil
}

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}
