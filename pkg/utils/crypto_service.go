// pkg/utils/crypto_service.go
package utils

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "io"
)

type CryptoService struct {
    Key []byte
}

func NewCryptoService(key []byte) *CryptoService {
    return &CryptoService{Key: key}
}

func (cs *CryptoService) Encrypt(plainText string) (string, error) {
    block, err := aes.NewCipher(cs.Key)
    if err != nil {
        return "", err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
    return base64.StdEncoding.EncodeToString(cipherText), nil
}

func (cs *CryptoService) Decrypt(encryptedText string) (string, error) {
    cipherData, err := base64.StdEncoding.DecodeString(encryptedText)
    if err != nil {
        return "", err
    }
    block, err := aes.NewCipher(cs.Key)
    if err != nil {
        return "", err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    if len(cipherData) < gcm.NonceSize() {
        return "", errors.New("ciphertext too short")
    }
    nonce := cipherData[:gcm.NonceSize()]
    cipherText := cipherData[gcm.NonceSize():]
    plainText, err := gcm.Open(nil, nonce, cipherText, nil)
    if err != nil {
        return "", err
    }
    return string(plainText), nil
}
