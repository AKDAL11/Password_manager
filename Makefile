SHELL := /bin/bash

.PHONY: build-android
build-android:
	source .env.android && cd ./cmd/main_desktop && fyne package --icon ../../Icon.png -os android --app-id com.akdal.passwordmanager && mv android.apk ../../password_manager.apk

.PHONY: build-desktop
build-desktop:
	source .env.desktop && cd ./cmd/main_desktop && go build -o ../../password_manager_desktop
