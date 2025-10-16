// i18n
package i18n

import (
    "bytes"
    "fmt"

    "gopkg.in/yaml.v3"
)

var translations map[string]map[string]string
var currentLang = "en"

func LoadLocale(lang string) error {
    var data []byte

    switch lang {
    case "ru":
        data = resourceRuYaml.Content()
    case "en":
        data = resourceEnYaml.Content()
    case "be":
        data = resourceBeYaml.Content()
    default:
        return fmt.Errorf("unknown locale: %s", lang)
    }

    var t map[string]map[string]string
    if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&t); err != nil {
        return fmt.Errorf("cannot parse yaml for %s: %w", lang, err)
    }

    translations = t
    currentLang = lang
    return nil
}

func T(key string) string {
    if translations == nil {
        return key
    }
    if val, ok := translations[key]; ok {
        if text, ok := val["other"]; ok {
            return text
        }
    }
    return key
}

func CurrentLang() string {
    return currentLang
}
