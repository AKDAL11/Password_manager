package model

// Полная структура пароля (для создания/обновления)
type Password struct {
	ID        int    `json:"id"`
	Service   string `json:"service"`
	Username  string `json:"username"`
	Link      string `json:"link"`
	Password  string `json:"password"`
	Category  string `json:"category"`
	CreatedAt string `json:"created_at"`
}

// Структура без поля Password (для публичного вывода)
type PasswordListItem struct {
	ID        int    `json:"id"`
	Service   string `json:"service"`
	Username  string `json:"username"`
	Link      string `json:"link"`
	Category  string `json:"category"`
	CreatedAt string `json:"created_at"`
}
