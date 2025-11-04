package utils

import (
	"encoding/base64"
	"strings"

	"password-manager/pkg/security"
)

type CryptoService struct {
    key []byte
}

func NewCryptoService(key []byte) *CryptoService {
    return &CryptoService{key: key}
}

func (c *CryptoService) Encrypt(plain string) (string, error) {
    enc, err := security.EncryptAESGCM(c.key, []byte(plain))
    if err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(enc), nil
}

func (c *CryptoService) Decrypt(encB64 string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encB64))
    if err != nil {
        return "", err
    }
    pt, err := security.DecryptAESGCM(c.key, data)
    if err != nil {
        return "", err
    }
    return string(pt), nil
}