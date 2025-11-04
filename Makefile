SHELL := /bin/bash

.PHONY: build-android
build-android:
  source .env.android && cd ./cmd/main_android && fyne package --icon ../../Icon.png -os android --app-id com.akdal.passwordmanager
  mv ./cmd/main_android/*.apk ./password_manager.apk

.PHONY: build-linux
build-linux:
  source .env.linux && cd ./cmd/main_desktop && go build -o ../../password_manager_linux

.PHONY: build-windows
build-windows:
    @powershell -Command "Get-Content .env.windows | ForEach-Object { $$parts = $$_ -split '='; if ($$parts.Length -eq 2) { [System.Environment]::SetEnvironmentVariable($$parts[0].Trim(), $$parts[1].Trim(), 'Process') } }; cd cmd/main_desktop; go build -o ../../password_manager.exe"
#go build -ldflags="-H=windowsgui" -o password_manager.exe