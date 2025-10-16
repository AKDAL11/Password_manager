package gui

import "password-manager/internal/app/model"

// extractSuggestions возвращает уникальные значения для автоподсказок
func extractSuggestions(passwords []model.PasswordListItem) (services, usernames, categories, links []string) {
    unique := func(values []string) []string {
        m := map[string]bool{}
        var result []string
        for _, v := range values {
            if v != "" && !m[v] {
                m[v] = true
                result = append(result, v)
            }
        }
        return result
    }

    var s, u, c, l []string
    for _, p := range passwords {
        s = append(s, p.Service)
        u = append(u, p.Username)
        c = append(c, p.Category)
        l = append(l, p.Link)
    }
    return unique(s), unique(u), unique(c), unique(l)
}
