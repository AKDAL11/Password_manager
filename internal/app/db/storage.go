package db

import (
    "password-manager/internal/app/model"
    "password-manager/pkg/utils"
)

type Storage interface {
    // Пароли
    GetAllPasswords() ([]model.PasswordListItem, error)
    GetPasswordByID(id string) (model.PasswordListItem, error)
    CreatePassword(p model.Password) (int64, string, error)
    UpdatePassword(id string, p model.Password) error
    DeletePassword(id string) error
    GetFilteredPasswords(service, username, category string) ([]model.PasswordListItem, error)
    GetEncryptedPasswordByID(id int) (string, error)
    Close() error

    // Meta (единый источник истины)
    HasMeta() bool
    SetCrypto(*utils.CryptoService)
}
