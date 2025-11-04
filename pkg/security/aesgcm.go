// aesgcm.go
package security

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "errors"
)

// EncryptAESGCM: шифрует plaintext, возвращает nonce||ciphertext.
func EncryptAESGCM(key []byte, plaintext []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    aead, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    nonce := make([]byte, aead.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return nil, err
    }
    ct := aead.Seal(nil, nonce, plaintext, nil)
    out := make([]byte, len(nonce)+len(ct))
    copy(out, nonce)
    copy(out[len(nonce):], ct)
    return out, nil
}

// DecryptAESGCM: принимает nonce||ciphertext, возвращает plaintext.
func DecryptAESGCM(key []byte, data []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    aead, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    n := aead.NonceSize()
    if len(data) < n {
        return nil, errors.New("invalid data")
    }
    return aead.Open(nil, data[:n], data[n:], nil)
}
