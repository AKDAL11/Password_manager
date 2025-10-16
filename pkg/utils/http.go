package utils

func JSONError(msg string) map[string]string {
    return map[string]string{"error": msg}
}
