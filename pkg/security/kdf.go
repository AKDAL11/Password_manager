// ksf.go
package security

import (
    "crypto/sha256"

    "golang.org/x/crypto/pbkdf2"
)

const (
    pbkdf2Iterations = 200_000
    derivedKeyLength = 32 // AES-256
)

// DeriveKey: ключ из пароля и соли через PBKDF2-SHA256.
func DeriveKey(password, salt []byte) []byte {
    return pbkdf2.Key(password, salt, pbkdf2Iterations, derivedKeyLength, sha256.New)
}
