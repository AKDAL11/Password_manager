package utils

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "io"
)

// CryptoService handles encryption and decryption
type CryptoService struct {
    Key []byte
}

// NewCryptoService creates a new CryptoService with a given key
func NewCryptoService(key []byte) *CryptoService {
    return &CryptoService{
        Key: key,
    }
}

// Encrypt encrypts plain text using AES
func (cs *CryptoService) Encrypt(plainText string) (string, error) {
    block, err := aes.NewCipher(cs.Key)
    if err != nil {
        return "", err
    }

    cipherText := make([]byte, aes.BlockSize+len(plainText))
    iv := cipherText[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }

    stream := cipher.NewCFBEncrypter(block, iv)
    stream.XORKeyStream(cipherText[aes.BlockSize:], []byte(plainText))

    return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt decrypts AES-encrypted text
func (cs *CryptoService) Decrypt(encryptedText string) (string, error) {
    cipherText, err := base64.StdEncoding.DecodeString(encryptedText)
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher(cs.Key)
    if err != nil {
        return "", err
    }

    if len(cipherText) < aes.BlockSize {
        return "", errors.New("ciphertext too short")
    }

    iv := cipherText[:aes.BlockSize]
    cipherText = cipherText[aes.BlockSize:]

    stream := cipher.NewCFBDecrypter(block, iv)
    stream.XORKeyStream(cipherText, cipherText)

    return string(cipherText), nil
}
