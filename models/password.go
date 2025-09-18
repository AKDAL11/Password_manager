package models

// Полная структура пароля (для создания/обновления)
type Password struct {
	ID        int    `json:"id"`
	Service   string `json:"service"`
	Username  string `json:"username"`
	Link      string `json:"link"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
}

// Структура без поля Password (для публичного вывода)
type PasswordListItem struct {
	ID        int    `json:"id"`
	Service   string `json:"service"`
	Username  string `json:"username"`
	Link      string `json:"link"`
	CreatedAt string `json:"created_at"`
}
