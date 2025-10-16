package utils

//validator.go
import "github.com/wagslane/go-password-validator"

func ValidatePasswordStrength(password string, minEntropy float64) error {
    return passwordvalidator.Validate(password, minEntropy)
}
