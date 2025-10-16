// internal/app/db/storage.go
package db

import "password-manager/internal/app/model"

type Storage interface {
    GetAllPasswords() ([]model.PasswordListItem, error)
    GetPasswordByID(id string) (model.PasswordListItem, error)
    CreatePassword(p model.Password) (int64, string, error)
    UpdatePassword(id string, p model.Password) error
    DeletePassword(id string) error
    GetEncryptedPassword(id string) (string, error)
    GetFilteredPasswords(service, username, category string) ([]model.PasswordListItem, error)
    Close() error

    HasMasterPassword() bool
    SaveMasterPassword(email, plain string) error
    VerifyMasterPassword(input string) error
    GetRecoveryEmail() (string, error)
}
