// pkg/applinux/applinux.go
//go:build !android

package applinux

import (
    "github.com/labstack/echo/v4"
    pmapp "password-manager/internal/app"
)

func InitApp(e *echo.Echo) *pmapp.App {
    return pmapp.InitApp(e, "passwords.db")
}
