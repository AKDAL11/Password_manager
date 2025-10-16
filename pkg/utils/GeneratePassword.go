package utils

import (
    "crypto/rand"
    "errors"
    "math/big"
    "strings"
)

func GeneratePassword(length int, useUpper, useLower, useDigits, useSymbols bool, exclude string) (string, error) {
    var charset string
    if useUpper {
        charset += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    }
    if useLower {
        charset += "abcdefghijklmnopqrstuvwxyz"
    }
    if useDigits {
        charset += "0123456789"
    }
    if useSymbols {
        charset += "!@#$%^&*()-_=+[]{}<>?/|"
    }

    for _, ch := range exclude {
        charset = strings.ReplaceAll(charset, string(ch), "")
    }

    if len(charset) == 0 {
        return "", errors.New("no characters available for generation")
    }

    password := make([]byte, length)
    for i := range password {
        index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
        if err != nil {
            return "", err
        }
        password[i] = charset[index.Int64()]
    }

    return string(password), nil
}
