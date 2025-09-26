package model

// Full password structure (used for creation/update)
type Password struct {
    ID        int    `json:"id"`
    Service   string `json:"service"`
    Username  string `json:"username"`
    Link      string `json:"link"`
    Password  string `json:"password"`
    Category  string `json:"category"`
    CreatedAt string `json:"created_at"`
}

// Structure without the Password field (used for public output)
type PasswordListItem struct {
    ID        int    `json:"id"`
    Service   string `json:"service"`
    Username  string `json:"username"`
    Link      string `json:"link"`
    Category  string `json:"category"`
    CreatedAt string `json:"created_at"`
    Password  string `json:"password"`
}
