// security.go
package security

import "crypto/rand"

// GenerateSalt: возвращает n случайных байт.
func GenerateSalt(n int) []byte {
    salt := make([]byte, n)
    _, _ = rand.Read(salt)
    return salt
}

// BytesEqual: константное сравнение байтов.
func BytesEqual(a, b []byte) bool {
    if len(a) != len(b) {
        return false
    }
    var v byte
    for i := 0; i < len(a); i++ {
        v |= a[i] ^ b[i]
    }
    return v == 0
}
