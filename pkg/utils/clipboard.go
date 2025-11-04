package utils

import (
	"fmt"

	"fyne.io/fyne/v2"
)

// CopyToClipboard: принимает base64-шифртекст и CryptoService,
// расшифровывает и кладёт в буфер обмена plain-пароль.
// Пустая строка очищает буфер.
func CopyToClipboard(encB64 string, cryptoSvc *CryptoService) error {
    if encB64 == "" {
        fyne.CurrentApp().Clipboard().SetContent("")
        return nil
    }
        
    plain, err := cryptoSvc.Decrypt(encB64)
    if err != nil {
        return fmt.Errorf("decrypt password: %w", err)
    }
    fyne.CurrentApp().Clipboard().SetContent(plain)
    return nil
}
